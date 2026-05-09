// sdk/go-sdk/client.go
// Nexus Hospitality Platform — Go SDK
// Supports: PMS, IPTV, IoT, AI, Billing APIs

package nexus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL = "https://api.nexus-platform.com"
	defaultVersion = "v1"
	userAgent      = "nexus-go-sdk/1.0.0"
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	apiKey     string
	tenantID   string
	userAgent  string

	// Service clients
	PMS      *PMSService
	IPTV     *IPTVService
	IoT      *IoTService
	AI       *AIService
	Billing  *BillingService
	Guest    *GuestService
}

type ClientOption func(*Client)

func WithAPIKey(key string) ClientOption {
	return func(c *Client) {
		c.apiKey = key
	}
}

func WithTenantID(tenant string) ClientOption {
	return func(c *Client) {
		c.tenantID = tenant
	}
}

func WithBaseURL(rawURL string) ClientOption {
	return func(c *Client) {
		u, _ := url.Parse(rawURL)
		c.baseURL = u
	}
}

func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

func NewClient(opts ...ClientOption) (*Client, error) {
	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  userAgent,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.PMS = &PMSService{client: c}
	c.IPTV = &IPTVService{client: c}
	c.IoT = &IoTService{client: c}
	c.AI = &AIService{client: c}
	c.Billing = &BillingService{client: c}
	c.Guest = &GuestService{client: c}

	return c, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	u, err := c.baseURL.Parse(fmt.Sprintf("/api/%s%s", defaultVersion, path))
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("X-Tenant-ID", c.tenantID)

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := &Response{
		Response: resp,
	}

	if resp.StatusCode >= 400 {
		return response, parseError(resp)
	}

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return response, err
		}
	}

	return response, nil
}

type Response struct {
	*http.Response
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d: %s - %s", e.Status, e.Code, e.Message)
}

func parseError(resp *http.Response) error {
	var err Error
	if err := json.NewDecoder(resp.Body).Decode(&err); err != nil {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	err.Status = resp.StatusCode
	return &err
}

// Pagination support
type ListOptions struct {
	Page    int    `url:"page,omitempty"`
	PerPage int    `url:"per_page,omitempty"`
	Sort    string `url:"sort,omitempty"`
	Order   string `url:"order,omitempty"`
}

type ListMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
