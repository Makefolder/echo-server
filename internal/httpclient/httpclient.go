// Project: Echo Server
// Description: Self-hosted communication server.
// Designed to be easily customisable and allow custom client implementations.
// Author: Makefolder
// Copyright (C) 2025, Artemii Fedotov <artemii.fedotov@tutamail.com>

package httpclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type HttpClient struct {
	httpClient *http.Client
	username   *string
	password   *string
}

func New(username, password *string, timeout time.Duration) *HttpClient {
	return &HttpClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		username: username,
		password: password,
	}
}

func (c *HttpClient) request(ctx context.Context, method, url string, headers map[string]string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.username != nil && c.password != nil {
		auth := base64.StdEncoding.EncodeToString([]byte(*c.username + ":" + *c.password))
		req.Header.Set("Authorization", "Basic "+auth)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.httpClient.Do(req)
}

func (c *HttpClient) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	return c.request(ctx, http.MethodGet, url, headers, nil)
}

func (c *HttpClient) Post(ctx context.Context, url string, headers map[string]string, body any) (*http.Response, error) {
	return c.request(ctx, http.MethodPost, url, headers, body)
}

func (c *HttpClient) Put(ctx context.Context, url string, headers map[string]string, body any) (*http.Response, error) {
	return c.request(ctx, http.MethodPut, url, headers, body)
}

func (c *HttpClient) Delete(ctx context.Context, url string, headers map[string]string) (*http.Response, error) {
	return c.request(ctx, http.MethodDelete, url, headers, nil)
}

func IsOK(res *http.Response) bool {
	return res != nil &&
		res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices
}

// LogResponse logs the response body
//   - log: zap sugarred logger
//   - body: res.Body, must be closed manually
func LogResponse(logger *zap.SugaredLogger, body io.Reader) error {
	if logger == nil || body == nil {
		return errors.New("invalid arguments")
	}

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Infof("response body: %s", bodyBytes)
	return nil
}
