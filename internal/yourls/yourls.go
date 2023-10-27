package yourls

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/devusSs/minio-link/internal/config/environment"
	"github.com/google/uuid"
)

// Wrapper for YOURLS API (basic)
type YOURLSClient struct {
	client    *http.Client
	baseURL   string
	signature string
}

// ShortenURL shortens a URL via YOURLS
func (c *YOURLSClient) ShortenURL(ctx context.Context, input string) (string, error) {
	_, err := checkURL(input)
	if err != nil {
		return "", fmt.Errorf("invalid input url: %w", err)
	}

	u, err := checkURL(fmt.Sprintf("%s/%s", c.baseURL, defaultAPIEndpoint))
	if err != nil {
		return "", fmt.Errorf("invalid base url: %w", err)
	}

	v := make(map[string]string)
	v["signature"] = c.signature
	v["action"] = "shorturl"
	v["format"] = "json"
	v["url"] = input
	v["title"] = "Uploaded using minio-yourls-uploader by devusSs"
	v["keyword"] = uuid.New().String()

	req, err := buildRequestWithContext(ctx, http.MethodPost, u.String(), createPostRequestBody(v))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var errRes shortenURLErrorResponse
		if err := unmarshalResponseToJSON(res, &errRes); err != nil {
			return "", fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return "", fmt.Errorf("failed to shorten url: %s", errRes.Message)
	}

	var shortenRes shortenURLResponse
	if err := unmarshalResponseToJSON(res, &shortenRes); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return shortenRes.Shorturl, nil
}

func (c *YOURLSClient) ExpandURL(ctx context.Context, input string) (string, error) {
	_, err := checkURL(input)
	if err != nil {
		return "", fmt.Errorf("invalid input url: %w", err)
	}

	u, err := checkURL(fmt.Sprintf("%s/%s", c.baseURL, defaultAPIEndpoint))
	if err != nil {
		return "", fmt.Errorf("invalid base url: %w", err)
	}

	v := make(map[string]string)
	v["signature"] = c.signature
	v["action"] = "expand"
	v["format"] = "json"
	v["shorturl"] = input

	req, err := buildRequestWithContext(ctx, http.MethodPost, u.String(), createPostRequestBody(v))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var errRes shortenURLErrorResponse
		if err := unmarshalResponseToJSON(res, &errRes); err != nil {
			return "", fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return "", fmt.Errorf("failed to shorten url: %s", errRes.Message)
	}

	var expandRes expandURLResponse
	if err := unmarshalResponseToJSON(res, &expandRes); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return expandRes.Longurl, nil
}

// NewClient creates a new YOURLSClient
func NewClient(dir string, debug bool, cfg *environment.EnvConfig) *YOURLSClient {
	return &YOURLSClient{
		client:    &http.Client{Timeout: 5 * time.Second},
		baseURL:   cfg.YourlsEndpoint,
		signature: cfg.YourlsSignatureKey,
	}
}

func checkURL(input string) (*url.URL, error) {
	u, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid input url: %w", err)
	}
	return u, nil
}

func createPostRequestBody(values map[string]string) io.Reader {
	v := url.Values{}
	for key, value := range values {
		v.Add(key, value)
	}
	return strings.NewReader(v.Encode())
}

func buildRequestWithContext(
	ctx context.Context,
	method string,
	url string,
	body io.Reader,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	return req, nil
}

func unmarshalResponseToJSON(res *http.Response, v interface{}) error {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	defer res.Body.Close()
	if err := json.Unmarshal(body, &v); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return nil
}

const (
	defaultAPIEndpoint string = "yourls-api.php"
)

type shortenURLErrorResponse struct {
	Status     string `json:"status"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	ErrorCode  string `json:"errorCode"`
	StatusCode string `json:"statusCode"`
}

type shortenURLResponse struct {
	URL struct {
		Keyword string `json:"keyword"`
		URL     string `json:"url"`
		Title   string `json:"title"`
		Date    string `json:"date"`
		IP      string `json:"ip"`
	} `json:"url"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	Title      string `json:"title"`
	Shorturl   string `json:"shorturl"`
	StatusCode int    `json:"statusCode"`
}

type expandURLResponse struct {
	Keyword    string `json:"keyword"`
	Shorturl   string `json:"shorturl"`
	Longurl    string `json:"longurl"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}
