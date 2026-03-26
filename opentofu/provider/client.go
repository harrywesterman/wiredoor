package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL *url.URL
	http    *http.Client
	token   string
}

func NewClient(cfg providerConfig) (*Client, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}

	parsed, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	transport := &http.Transport{}
	if cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402
	}

	return &Client{
		baseURL: parsed,
		http: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		token: cfg.Token,
	}, nil
}

func (c *Client) Authenticate(username, password string) error {
	if c.token != "" {
		return nil
	}

	resp, err := c.doRaw(http.MethodPost, "/auth/login", map[string]string{
		"username": username,
		"password": password,
	}, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseError(resp)
	}

	var auth authResponse
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return err
	}

	if auth.Token == "" {
		return fmt.Errorf("authentication succeeded without token")
	}

	c.token = auth.Token
	return nil
}

func (c *Client) GetConfig() (*publicConfig, error) {
	var out publicConfig
	if err := c.do(http.MethodGet, "/config", nil, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetNode(id int64) (*node, error) {
	var out node
	if err := c.do(http.MethodGet, fmt.Sprintf("/nodes/%d", id), nil, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateNode(payload map[string]interface{}) (*node, error) {
	var out node
	if err := c.do(http.MethodPost, "/nodes", payload, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateNode(id int64, payload map[string]interface{}) (*node, error) {
	var out node
	if err := c.do(http.MethodPatch, fmt.Sprintf("/nodes/%d", id), payload, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteNode(id int64) error {
	return c.do(http.MethodDelete, fmt.Sprintf("/nodes/%d", id), nil, true, nil)
}

func (c *Client) GetHTTPService(nodeID, serviceID int64) (*httpService, error) {
	var out httpService
	if err := c.do(http.MethodGet, fmt.Sprintf("/services/%d/http/%d", nodeID, serviceID), nil, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateHTTPService(nodeID int64, payload map[string]interface{}) (*httpService, error) {
	var out httpService
	if err := c.do(http.MethodPost, fmt.Sprintf("/services/%d/http", nodeID), payload, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateHTTPService(nodeID, serviceID int64, payload map[string]interface{}) (*httpService, error) {
	var out httpService
	if err := c.do(http.MethodPatch, fmt.Sprintf("/services/%d/http/%d", nodeID, serviceID), payload, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteHTTPService(nodeID, serviceID int64) error {
	return c.do(http.MethodDelete, fmt.Sprintf("/services/%d/http/%d", nodeID, serviceID), nil, true, nil)
}

func (c *Client) GetTCPService(nodeID, serviceID int64) (*tcpService, error) {
	var out tcpService
	if err := c.do(http.MethodGet, fmt.Sprintf("/services/%d/tcp/%d", nodeID, serviceID), nil, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateTCPService(nodeID int64, payload map[string]interface{}) (*tcpService, error) {
	var out tcpService
	if err := c.do(http.MethodPost, fmt.Sprintf("/services/%d/tcp", nodeID), payload, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateTCPService(nodeID, serviceID int64, payload map[string]interface{}) (*tcpService, error) {
	var out tcpService
	if err := c.do(http.MethodPatch, fmt.Sprintf("/services/%d/tcp/%d", nodeID, serviceID), payload, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteTCPService(nodeID, serviceID int64) error {
	return c.do(http.MethodDelete, fmt.Sprintf("/services/%d/tcp/%d", nodeID, serviceID), nil, true, nil)
}

func (c *Client) GetDomain(id int64) (*domain, error) {
	var out domain
	if err := c.do(http.MethodGet, fmt.Sprintf("/domains/%d", id), nil, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateDomain(payload map[string]interface{}) (*domain, error) {
	var out domain
	if err := c.do(http.MethodPost, "/domains", payload, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateDomain(id int64, payload map[string]interface{}) (*domain, error) {
	var out domain
	if err := c.do(http.MethodPatch, fmt.Sprintf("/domains/%d", id), payload, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteDomain(id int64) error {
	return c.do(http.MethodDelete, fmt.Sprintf("/domains/%d", id), nil, true, nil)
}

func (c *Client) GetPAT(nodeID, patID int64) (*pat, error) {
	var out pat
	if err := c.do(http.MethodGet, fmt.Sprintf("/nodes/%d/pats?id=%d", nodeID, patID), nil, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreatePAT(nodeID int64, payload map[string]interface{}) (*pat, error) {
	var out pat
	if err := c.do(http.MethodPost, fmt.Sprintf("/nodes/%d/pats", nodeID), payload, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeletePAT(nodeID, patID int64) error {
	return c.do(http.MethodDelete, fmt.Sprintf("/nodes/%d/pats/%d", nodeID, patID), nil, true, nil)
}

func (c *Client) RevokePAT(nodeID, patID int64) (*pat, error) {
	var out pat
	if err := c.do(http.MethodPatch, fmt.Sprintf("/nodes/%d/pats/%d/revoke", nodeID, patID), nil, true, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) do(method, rawPath string, payload interface{}, auth bool, out interface{}) error {
	resp, err := c.doRaw(method, rawPath, payload, auth)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseError(resp)
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) doRaw(method, rawPath string, payload interface{}, auth bool) (*http.Response, error) {
	u := *c.baseURL
	u.Path = path.Join(strings.TrimSuffix(c.baseURL.Path, "/"), rawPath)

	var body io.Reader
	if payload != nil {
		buf, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(buf)
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if auth && c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.http.Do(req)
}

func responseError(resp *http.Response) error {
	b, _ := io.ReadAll(resp.Body)
	message := strings.TrimSpace(string(b))
	if message == "" {
		message = resp.Status
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: %s", message)
	}
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("forbidden: %s", message)
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found: %s", message)
	}
	return fmt.Errorf("%s: %s", resp.Status, message)
}

func int64Value(v interface{}) int64 {
	switch t := v.(type) {
	case int:
		return int64(t)
	case int64:
		return t
	case float64:
		return int64(t)
	case string:
		n, _ := strconv.ParseInt(t, 10, 64)
		return n
	default:
		return 0
	}
}

