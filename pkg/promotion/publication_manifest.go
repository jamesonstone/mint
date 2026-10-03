package promotion

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ReleaseManifest excludes mutable publication status and is stable across
// publication retries. Its deployment evidence is already durable in the journal.
func ReleaseManifest(i Intent) ([]byte, error) {
	return json.MarshalIndent(struct {
		IntentID   string    `json:"intent_id"`
		MergeSHA   string    `json:"merge_sha"`
		BaselineID string    `json:"baseline_id"`
		Candidate  Candidate `json:"candidate"`
		RunID      int64     `json:"deployment_run_id"`
		URL        string    `json:"deployment_url"`
		Notes      string    `json:"notes"`
	}{i.ID, i.MergeSHA, i.BaselineID, i.Candidate, i.DeploymentRunID, i.DeploymentURL, i.Notes}, "", "  ")
}
func (c Client) attachReleaseManifest(ctx context.Context, id int64, i Intent, upload bool) error {
	data, err := ReleaseManifest(i)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	digest := fmt.Sprintf("sha256:%x", sum)
	for page := 1; ; page++ {
		var assets []struct {
			Name, Digest string
			Size         int64
		}
		status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("releases/%d/assets?per_page=100&page=%d", id, page)), nil, &assets)
		if err != nil {
			return err
		}
		if status != 200 {
			return fmt.Errorf("release assets unavailable")
		}
		for _, asset := range assets {
			if asset.Name == "mint-production.json" {
				if asset.Size != int64(len(data)) || asset.Digest != digest {
					return fmt.Errorf("existing release manifest differs from deployed intent")
				}
				return nil
			}
		}
		if len(assets) < 100 {
			break
		}
	}
	if !upload {
		return fmt.Errorf("production release manifest is absent")
	}
	endpoint := "https://uploads.github.com/" + c.repoPath(fmt.Sprintf("releases/%d/assets?name=mint-production.json", id))
	// Custom API servers use their own uploads endpoint (also useful for loopback tests).
	if c.APIURL != "https://api.github.com" {
		endpoint = strings.TrimRight(c.APIURL, "/") + "/" + c.repoPath(fmt.Sprintf("releases/%d/assets?name=mint-production.json", id))
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || (parsed.Scheme != "https" && parsed.Hostname() != "127.0.0.1") {
		return fmt.Errorf("unsafe release upload endpoint")
	}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	hc := &http.Client{Timeout: 45 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	if c.HTTP != nil {
		hc.Transport = c.HTTP.Transport
	}
	res, err := hc.Do(req)
	if err != nil {
		return fmt.Errorf("release manifest upload failed")
	}
	defer func() { _ = res.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 1<<20))
	if res.StatusCode != 201 {
		return fmt.Errorf("release manifest upload returned %d; retry publication only", res.StatusCode)
	}
	return nil
}
