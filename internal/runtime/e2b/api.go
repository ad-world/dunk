package e2b

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dunk/internal/runtime"
)

const (
	platformBaseURL = "https://api.e2b.app"
	sandboxBaseURL  = "https://sandbox.e2b.app"
	envdPort        = "49983"
)

type apiClient struct {
	apiKey string
	http   *http.Client
}

type providerState struct {
	SandboxID       string `json:"sandbox_id"`
	EnvdAccessToken string `json:"envd_access_token,omitempty"`
}

type sandboxResp struct {
	SandboxID       string `json:"sandboxID"`
	EnvdAccessToken string `json:"envdAccessToken"`
}

func newAPIClient(apiKey string) apiClient {
	return apiClient{apiKey: apiKey, http: &http.Client{Timeout: 60 * time.Second}}
}

func (c apiClient) createSandbox(ctx context.Context, template string, timeoutSeconds int, metadata map[string]string) (sandboxResp, error) {
	body := map[string]any{
		"templateID": template,
		"timeout":    timeoutSeconds,
		"metadata":   metadata,
		"secure":     false,
	}
	return c.platform(ctx, http.MethodPost, platformBaseURL+"/sandboxes", body, http.StatusCreated)
}

func (c apiClient) connectSandbox(ctx context.Context, id string, timeoutSeconds int) (sandboxResp, error) {
	return c.platform(ctx, http.MethodPost, platformBaseURL+"/sandboxes/"+url.PathEscape(id)+"/connect", map[string]any{"timeout": timeoutSeconds}, 0)
}

func (c apiClient) deleteSandbox(ctx context.Context, id string) error {
	if c.apiKey == "" {
		return errors.New("E2B_API_KEY is not set")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, platformBaseURL+"/sandboxes/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotFound {
		return nil
	}
	b, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("e2b delete sandbox: %s: %s", resp.Status, strings.TrimSpace(string(b)))
}

func (c apiClient) uploadFile(ctx context.Context, ws *runtime.Workspace, local, remote string) error {
	f, err := os.Open(local)
	if err != nil {
		return err
	}
	defer f.Close()

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("file", filepath.Base(local))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, f); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sandboxBaseURL+"/files?path="+url.QueryEscape(remote), &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("E2b-Sandbox-Id", ws.ID)
	req.Header.Set("E2b-Sandbox-Port", envdPort)
	if tok := envdToken(ws); tok != "" {
		req.Header.Set("X-Access-Token", tok)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("e2b upload %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	return nil
}

func (c apiClient) platform(ctx context.Context, method, endpoint string, body any, want int) (sandboxResp, error) {
	if c.apiKey == "" {
		return sandboxResp{}, errors.New("E2B_API_KEY is not set")
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return sandboxResp{}, err
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, &buf)
	if err != nil {
		return sandboxResp{}, err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return sandboxResp{}, err
	}
	defer resp.Body.Close()
	if want == 0 {
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			b, _ := io.ReadAll(resp.Body)
			return sandboxResp{}, fmt.Errorf("e2b api %s: %s", resp.Status, strings.TrimSpace(string(b)))
		}
	} else if resp.StatusCode != want {
		b, _ := io.ReadAll(resp.Body)
		return sandboxResp{}, fmt.Errorf("e2b api %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	var out sandboxResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return sandboxResp{}, err
	}
	if out.SandboxID == "" {
		return sandboxResp{}, errors.New("e2b response missing sandboxID")
	}
	return out, nil
}

func envdToken(ws *runtime.Workspace) string {
	var st providerState
	_ = json.Unmarshal(ws.ProviderState, &st)
	return st.EnvdAccessToken
}
