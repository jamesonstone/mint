package promotion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestPolicyPermitsExactlyOneTargetUnderOriginalAuthority(t *testing.T) {
	before, err := ParsePolicy([]byte(environmentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	after, err := ParsePolicy([]byte(strings.Replace(environmentPolicyYAML, "target: v1.2.3", "target: v1.2.4", 1)))
	if err != nil {
		t.Fatal(err)
	}
	environment, err := PolicyRequestEnvironment(before, after)
	if err != nil || environment != "production" {
		t.Fatal(environment, err)
	}
	if before.Environments["production"].Target != "v1.2.3" || after.Environments["production"].Target != "v1.2.4" {
		t.Fatal("authority comparison mutated request policies")
	}
	if _, err := PolicyRequestEnvironment(before, before); err == nil {
		t.Fatal("unchanged YAML triggered deployment")
	}
}
func TestRequestPolicyCannotWeakenOrReplaceItsOwnAuthority(t *testing.T) {
	before, err := ParsePolicy([]byte(environmentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(environmentPolicyYAML, "target: v1.2.3", "target: v1.2.4", 1)
	cases := map[string]string{
		"downgrade":         strings.Replace(changed, "deploy: reviewed", "deploy: automatic", 1),
		"workflow":          strings.Replace(changed, ".github/workflows/deploy.yaml", ".github/workflows/other.yaml", 1),
		"checks":            strings.Replace(changed, "required_checks: [native]", "required_checks: [weaker]", 1),
		"operator":          strings.Replace(changed, "human_login: jamesonstone", "human_login: intruder", 1),
		"config":            strings.Replace(changed, "sha256:"+strings.Repeat("a", 64), "sha256:"+strings.Repeat("b", 64), 1),
		"unrelated-env":     strings.Replace(changed, "deploy: automatic", "deploy: manual", 1),
		"multiple-requests": strings.Replace(changed, "follow: latest", "target: v1.0.0", 1),
		"publication":       strings.Replace(changed, "publish: true", "publish: false", 1),
		"default":           changed + "default_environment: dev\n",
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			after, err := ParsePolicy([]byte(data))
			if err != nil {
				t.Fatal("request fixture should be valid independently", err)
			}
			if _, err := PolicyRequestEnvironment(before, after); err == nil {
				t.Fatal("request weakened or replaced original policy")
			}
		})
	}
	removed := before
	removed.Environments = map[string]EnvironmentPolicy{"production": before.Environments["production"]}
	if _, err := PolicyRequestEnvironment(before, removed); err == nil {
		t.Fatal("environment removal approved deployment")
	}
	added := before
	added.Environments = map[string]EnvironmentPolicy{}
	for name, env := range before.Environments {
		added.Environments[name] = env
	}
	added.Environments["extra"] = added.Environments["dev"]
	if _, err := PolicyRequestEnvironment(before, added); err == nil {
		t.Fatal("new policy environment approved its own deployment")
	}
}
func TestRequestPolicyAcceptsExplicitResumeAndRejectsLegacyAuthority(t *testing.T) {
	before, err := ParsePolicy([]byte(environmentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	data := strings.Replace(environmentPolicyYAML, "target: v1.2.3", "follow: latest\n    operation: resume", 1)
	after, err := ParsePolicy([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	environment, err := PolicyRequestEnvironment(before, after)
	if err != nil || environment != "production" {
		t.Fatal(environment, err)
	}
	before.Schema = 1
	if _, err := PolicyRequestEnvironment(before, after); err == nil {
		t.Fatal("legacy policy approved schema2 request")
	}
	invalid := strings.Replace(environmentPolicyYAML, "scope: shared", "scope: local", 1)
	if _, err := ParsePolicy([]byte(invalid)); err == nil {
		t.Fatal("shared deployment accepted machine-local prerequisite")
	}
}
func TestRequestAuthorityUsesExactMergeParent(t *testing.T) {
	for _, valid := range []bool{true, false} {
		t.Run(map[bool]string{true: "exact", false: "foreign-response"}[valid], func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/repos/owner/repo/git/commits/"+sha(4) || r.Method != "GET" {
					t.Error(r.Method, r.URL.Path)
				}
				identity := sha(4)
				if !valid {
					identity = sha(3)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"sha": identity, "parents": []any{map[string]string{"sha": sha(2)}, map[string]string{"sha": sha(1)}}})
			}))
			defer server.Close()
			c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
			parent, err := c.PolicyParent(context.Background(), sha(4))
			if valid && (err != nil || parent != sha(2)) {
				t.Fatal(parent, err)
			}
			if !valid && err == nil {
				t.Fatal("foreign commit response accepted")
			}
		})
	}
	if _, err := (Client{}).PolicyParent(context.Background(), "main"); err == nil {
		t.Fatal("mutable merge ref accepted")
	}
}
