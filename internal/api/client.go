package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/gregPerlinLi/CertVaultCLIX/internal/version"
	"golang.org/x/net/publicsuffix"
)

// ErrUnauthorized is returned when the server returns HTTP 401 or API code 401,
// indicating the session has expired and the user must log in again.
var ErrUnauthorized = errors.New("session expired: please log in again")

const defaultTimeout = 30 * time.Second

// Client is the CertVault API HTTP client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	session    string
}

// NewClient creates a new API client.
func NewClient(baseURL string) *Client {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
			Jar:     jar,
		},
	}
}

// SetSession sets the JSESSIONID cookie on the client.
func (c *Client) SetSession(session string) {
	c.session = session
}

// GetSession returns the current JSESSIONID value.
func (c *Client) GetSession() string {
	return c.session
}

// SetBaseURL updates the base URL.
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// GetBaseURL returns the base URL.
func (c *Client) GetBaseURL() string {
	return c.baseURL
}

func (c *Client) userAgent() string {
	return fmt.Sprintf("Mozilla/5.0 (Linux) CertVaultCLIX/%s", version.Version)
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent())
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.session != "" {
		req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: c.session})
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Extract JSESSIONID from response cookies
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			c.session = cookie.Value
		}
	}

	return resp, nil
}

func (c *Client) get(ctx context.Context, path string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

func (c *Client) put(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.do(ctx, http.MethodPut, path, body)
}

func (c *Client) patch(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	return c.do(ctx, http.MethodPatch, path, body)
}

func (c *Client) delete(ctx context.Context, path string) (*http.Response, error) {
	return c.do(ctx, http.MethodDelete, path, nil)
}

// decodeResponse decodes a JSON response into the given type.
func decodeResponse[T any](resp *http.Response) (*ResultVO[T], error) {
	defer resp.Body.Close()
	// Treat HTTP 401 as session-expired regardless of body
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	// HTTP 204 No Content — return a zero-value success result (e.g. empty PageDTO).
	// Code is normalised to 200 so all callers can treat it uniformly as a success.
	if resp.StatusCode == http.StatusNoContent {
		var zero T
		return &ResultVO[T]{Code: 200, Data: zero}, nil
	}
	var result ResultVO[T]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	// Some API implementations return 401 in the body code field
	if result.Code == 401 {
		return nil, ErrUnauthorized
	}
	if result.Code != 200 && result.Code != 0 {
		return &result, fmt.Errorf("API error %d: %s", result.Code, result.Msg)
	}
	return &result, nil
}
