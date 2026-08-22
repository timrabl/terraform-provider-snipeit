// Package client implements a minimal HTTP client for the Snipe-IT REST API
// (api/v1).
//
// API quirks this client papers over:
//   - Mutations (POST/PUT/PATCH/DELETE) and error cases respond with HTTP 200
//     and an envelope {"status": "success"|"error", "messages": ..., "payload": ...}.
//     Success/failure must be derived from the envelope, not the HTTP status.
//   - GET on a single object returns the raw object without an envelope, but a
//     missing object still yields HTTP 200 with an error envelope.
//   - Create responses carry only a partial payload (submitted fields + id +
//     timestamps), so callers must re-GET after create to obtain the full object.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// ErrNotFound is returned when the API reports that the requested object does
// not exist.
var ErrNotFound = errors.New("snipe-it: not found")

// Client talks to a Snipe-IT instance's REST API.
type Client struct {
	baseURL    string // e.g. https://snipeit.example.com/api/v1
	token      string
	httpClient *http.Client
	userAgent  string
	// retryDelay computes the backoff before retry number attempt+1; it is a
	// field only so unit tests can eliminate the wait.
	retryDelay func(attempt int) time.Duration
}

// defaultRetryDelay backs off progressively; the API limiter window is per
// minute.
func defaultRetryDelay(attempt int) time.Duration {
	return time.Duration(attempt+1) * 2 * time.Second
}

// Config holds the settings needed to construct a Client.
type Config struct {
	// URL is the base URL of the Snipe-IT instance, without the /api/v1 suffix.
	URL string
	// Token is the personal API (Bearer) token.
	Token string
	// Insecure disables TLS certificate verification (self-signed dev instances).
	Insecure bool
	// UserAgent is sent with every request.
	UserAgent string
	// Timeout for individual HTTP requests; defaults to 30s.
	Timeout time.Duration
}

// New constructs a Client from cfg.
func New(cfg Config) (*Client, error) {
	base := strings.TrimRight(cfg.URL, "/")
	if base == "" {
		return nil, fmt.Errorf("snipe-it: URL must not be empty")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("snipe-it: token must not be empty")
	}
	if !strings.HasSuffix(base, "/api/v1") {
		base += "/api/v1"
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // explicit opt-in for dev instances
	}

	return &Client{
		baseURL:   base,
		token:     cfg.Token,
		userAgent: cfg.UserAgent,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		retryDelay: defaultRetryDelay,
	}, nil
}

// envelope is the mutation/error response wrapper used by Snipe-IT.
type envelope struct {
	Status   string          `json:"status"`
	Messages json.RawMessage `json:"messages"`
	Payload  json.RawMessage `json:"payload"`
}

// APIError is a non-success envelope returned by the API.
type APIError struct {
	StatusCode int
	Messages   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("snipe-it API error (HTTP %d): %s", e.StatusCode, e.Messages)
}

// flattenMessages renders the polymorphic "messages" field (string, or map of
// field name -> list of validation errors) into a single readable string.
func flattenMessages(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var m map[string][]string
	if err := json.Unmarshal(raw, &m); err == nil {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var parts []string
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s: %s", k, strings.Join(m[k], "; ")))
		}
		return strings.Join(parts, ", ")
	}
	return string(raw)
}

func isNotFoundMessage(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "not found") || strings.Contains(lower, "does not exist")
}

// maxRetries is the number of attempts for requests rejected by the API rate
// limiter (HTTP 429) before giving up.
const maxRetries = 8

func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, int, error) {
	var payload []byte
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("snipe-it: encoding request body: %w", err)
		}
		payload = buf
	}

	var lastData []byte
	var lastCode int
	for attempt := 0; attempt < maxRetries; attempt++ {
		data, code, err := c.doOnce(ctx, method, path, payload)
		if err != nil {
			return nil, code, err
		}
		if code != http.StatusTooManyRequests {
			return data, code, nil
		}
		lastData, lastCode = data, code

		delay := c.retryDelay(attempt)
		select {
		case <-ctx.Done():
			return nil, code, ctx.Err()
		case <-time.After(delay):
		}
	}
	return lastData, lastCode, nil
}

func (c *Client) doOnce(ctx context.Context, method, path string, payload []byte) ([]byte, int, error) {
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, 0, fmt.Errorf("snipe-it: building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("snipe-it: %s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("snipe-it: reading response: %w", err)
	}
	return data, resp.StatusCode, nil
}

// Get fetches a single object (raw, unwrapped) into out. Returns ErrNotFound
// if the API reports the object missing.
func (c *Client) Get(ctx context.Context, path string, out any) error {
	data, code, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	if code == http.StatusNotFound {
		return ErrNotFound
	}

	// A GET can still return an error envelope with HTTP 200.
	var env envelope
	if err := json.Unmarshal(data, &env); err == nil && env.Status == "error" {
		msg := flattenMessages(env.Messages)
		if isNotFoundMessage(msg) {
			return ErrNotFound
		}
		return &APIError{StatusCode: code, Messages: msg}
	}
	if code >= 400 {
		return &APIError{StatusCode: code, Messages: strings.TrimSpace(string(data))}
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("snipe-it: decoding GET %s response: %w", path, err)
	}
	return nil
}

// mutate sends a write request and unwraps the response envelope. The payload
// (if any and if out != nil) is decoded into out.
func (c *Client) mutate(ctx context.Context, method, path string, body, out any) error {
	data, code, err := c.do(ctx, method, path, body)
	if err != nil {
		return err
	}

	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		if code >= 400 {
			return &APIError{StatusCode: code, Messages: strings.TrimSpace(string(data))}
		}
		return fmt.Errorf("snipe-it: decoding %s %s envelope: %w", method, path, err)
	}
	if env.Status != "success" {
		msg := flattenMessages(env.Messages)
		if isNotFoundMessage(msg) {
			return ErrNotFound
		}
		return &APIError{StatusCode: code, Messages: msg}
	}
	if out != nil && len(env.Payload) > 0 && string(env.Payload) != "null" {
		if err := json.Unmarshal(env.Payload, out); err != nil {
			return fmt.Errorf("snipe-it: decoding %s %s payload: %w", method, path, err)
		}
	}
	return nil
}

// Post creates an object and decodes the (partial) payload into out.
func (c *Client) Post(ctx context.Context, path string, body, out any) error {
	return c.mutate(ctx, http.MethodPost, path, body, out)
}

// Patch partially updates an object.
func (c *Client) Patch(ctx context.Context, path string, body, out any) error {
	return c.mutate(ctx, http.MethodPatch, path, body, out)
}

// Put replaces an object.
func (c *Client) Put(ctx context.Context, path string, body, out any) error {
	return c.mutate(ctx, http.MethodPut, path, body, out)
}

// Delete removes an object. Deleting an already-missing object returns
// ErrNotFound, which callers typically treat as success.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.mutate(ctx, http.MethodDelete, path, nil, nil)
}
