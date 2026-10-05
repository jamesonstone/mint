package promotion

import (
	"context"
	"fmt"
	"reflect"
)

// PolicyRequestEnvironment allows request fields in exactly one environment.
// All authority-bearing policy must equal the trusted pre-change document.
func PolicyRequestEnvironment(before, after Policy) (string, error) {
	if before.Schema != 2 || after.Schema != 2 {
		return "", fmt.Errorf("deployment requests require schema 2 policies")
	}
	left, right := before, after
	left.Environments = make(map[string]EnvironmentPolicy, len(before.Environments))
	right.Environments = make(map[string]EnvironmentPolicy, len(after.Environments))
	changed := ""
	for name, env := range before.Environments {
		next, ok := after.Environments[name]
		if !ok {
			return "", fmt.Errorf("a deployment request cannot remove an environment")
		}
		if env.Target != next.Target || env.Follow != next.Follow || env.Operation != next.Operation || env.Reason != next.Reason {
			if changed != "" {
				return "", fmt.Errorf("one deployment request per PR")
			}
			changed = name
		}
		env.Target, env.Follow, env.Operation, env.Reason = "", "", "", ""
		next.Target, next.Follow, next.Operation, next.Reason = "", "", "", ""
		left.Environments[name], right.Environments[name] = env, next
	}
	if len(before.Environments) != len(after.Environments) || !reflect.DeepEqual(left, right) {
		return "", fmt.Errorf("deployment requests cannot change their own trusted policy")
	}
	if changed == "" {
		return "", fmt.Errorf("policy-only change does not request a deployment")
	}
	return changed, nil
}

// PolicyParent reads the actual first parent of the exact merged commit rather
// than a mutable default-branch snapshot supplied in an event.
func (c Client) PolicyParent(ctx context.Context, sha string) (string, error) {
	if !shaPattern.MatchString(sha) {
		return "", fmt.Errorf("exact merge identity required")
	}
	var commit struct {
		SHA     string
		Parents []struct{ SHA string }
	}
	status, err := c.request(ctx, "GET", c.repoPath("git/commits/"+sha), nil, &commit)
	if err != nil {
		return "", err
	}
	if status != 200 || commit.SHA != sha || len(commit.Parents) == 0 || !shaPattern.MatchString(commit.Parents[0].SHA) {
		return "", fmt.Errorf("merge parent unavailable")
	}
	return commit.Parents[0].SHA, nil
}

// CheckPolicyRequestSource compares authority before the entire reviewed PR,
// as well as before its final merge commit. Rebase merges must not smuggle an
// earlier authority change into the apparent parent of the final target edit.
func (c Client) CheckPolicyRequestSource(ctx context.Context,pr PullRequest,path string,before Policy,environment string) error {
 commits,err:=c.hotfixPRCommits(ctx,pr.Number)
 if err!=nil { return err }
 if len(commits)==0 { return fmt.Errorf("reviewed policy request has no provable source history") }
 var first struct { SHA string; Parents []struct{SHA string} }
 status,err:=c.request(ctx,"GET",c.repoPath("git/commits/"+commits[0]),nil,&first)
 if err!=nil { return err }
 if status!=200 || first.SHA!=commits[0] || len(first.Parents)!=1 || !shaPattern.MatchString(first.Parents[0].SHA) { return fmt.Errorf("ambiguous policy request base; create a fresh target-only PR from the current default branch") }
 oldData,err:=c.ReadFile(ctx,path,first.Parents[0].SHA)
 if err!=nil { return err }
 newData,err:=c.ReadFile(ctx,path,pr.Head.SHA)
 if err!=nil { return err }
 sourceBefore,err:=ParsePolicy(oldData)
 if err!=nil { return err }
 sourceAfter,err:=ParsePolicy(newData)
 if err!=nil { return err }
 name,err:=PolicyRequestEnvironment(sourceBefore,sourceAfter)
 if err!=nil || name!=environment { return fmt.Errorf("whole reviewed PR changes deployment authority or a different request") }
 left,right:=sourceBefore,before
 left.Environments=make(map[string]EnvironmentPolicy,len(sourceBefore.Environments))
 right.Environments=make(map[string]EnvironmentPolicy,len(before.Environments))
 for name,env:=range sourceBefore.Environments { env.Target,env.Follow,env.Operation,env.Reason="","","","";left.Environments[name]=env }
 for name,env:=range before.Environments { env.Target,env.Follow,env.Operation,env.Reason="","","","";right.Environments[name]=env }
 if !reflect.DeepEqual(left,right) { return fmt.Errorf("policy authority changed within or since the reviewed PR; create and review a fresh target-only request") }
 return nil
}
