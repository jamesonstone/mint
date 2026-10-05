package promotion

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrManifestUnavailable = errors.New("producer manifest unavailable")

// RunManifest downloads the unique manifest from the server-attested producer
// run. A local JSON file is never evidence of the artifact that run produced.
func (c Client) RunManifest(ctx context.Context, runID int64, name string, out any) error {
	var selected int64
	var digest string
	for page := 1; ; page++ {
		var list struct {
			Total     int `json:"total_count"`
			Artifacts []struct {
				ID      int64
				Name    string
				Expired bool
				Digest  string
			}
		}
		status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("actions/runs/%d/artifacts?per_page=100&page=%d", runID, page)), nil, &list)
		if err != nil {
			return err
		}
		if status != 200 {
			return fmt.Errorf("producer artifacts unavailable")
		}
		for _, a := range list.Artifacts {
			if a.Name == name {
				if a.Expired {
					return ErrManifestUnavailable
				}
				if selected != 0 {
					return fmt.Errorf("producer manifest is ambiguous or expired")
				}
				selected = a.ID
				digest = a.Digest
			}
		}
		if page*100 >= list.Total {
			break
		}
	}
	if selected == 0 {
		return ErrManifestUnavailable
	}
	if !digestPattern.MatchString(digest) {
		return fmt.Errorf("unique digest-attested producer manifest required")
	}
	archive, err := c.downloadArchive(ctx, selected)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(archive)
	if fmt.Sprintf("sha256:%x", sum) != digest {
		return fmt.Errorf("producer archive digest mismatch")
	}
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return err
	}
	if len(zr.File) != 1 || zr.File[0].Name != name+".json" || zr.File[0].UncompressedSize64 > 1<<20 {
		return fmt.Errorf("unexpected producer manifest archive")
	}
	file, err := zr.File[0].Open()
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	decoder := json.NewDecoder(io.LimitReader(file, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return fmt.Errorf("trailing producer manifest content")
	}
	return nil
}
func (c Client) downloadArchive(ctx context.Context, id int64) ([]byte, error) {
	endpoint := strings.TrimRight(c.APIURL, "/") + "/" + c.repoPath(fmt.Sprintf("actions/artifacts/%d/zip", id))
	parsed, err := url.Parse(endpoint)
	if err != nil || (parsed.Scheme != "https" && parsed.Hostname() != "127.0.0.1") {
		return nil, fmt.Errorf("unsafe artifact API")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	hc := &http.Client{Timeout: 45 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	if c.HTTP != nil {
		hc.Transport = c.HTTP.Transport
	}
	response, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("producer archive request failed")
	}
	if response.StatusCode == 302 {
		location, err := url.Parse(response.Header.Get("Location"))
		_ = response.Body.Close()
		if err != nil || location.Scheme != "https" || location.User != nil || location.Host == "" {
			return nil, fmt.Errorf("unsafe signed artifact URL")
		}
		// This URL is returned by the authenticated GitHub API. No credentials are
		// forwarded to storage and no second redirect is followed.
		req, err = http.NewRequestWithContext(ctx, "GET", location.String(), nil)
		if err != nil {
			return nil, err
		}
		response, err = hc.Do(req)
		if err != nil {
			return nil, fmt.Errorf("producer archive download failed")
		}
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("producer archive unavailable: %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 8<<20+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 8<<20 {
		return nil, fmt.Errorf("producer archive exceeds manifest limit")
	}
	return data, nil
}

// DeploymentManifest is emitted only after the application adapter verifies the
// exact artifact and running production identity.
type DeploymentManifest struct {
	Environment    string    `json:"environment,omitempty"`
	Artifact       *Artifact `json:"artifact,omitempty"`
	Outcome        string    `json:"outcome,omitempty"`
	IntentID       string    `json:"intent_id"`
	SourceSHA      string    `json:"source_sha"`
	ArtifactDigest string    `json:"artifact_digest"`
	Configuration  string    `json:"configuration_sha256"`
	Verified       bool      `json:"verified"`
}
