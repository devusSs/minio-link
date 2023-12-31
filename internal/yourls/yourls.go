package yourls

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/devusSs/minio-link/internal/config/environment"
	"github.com/devusSs/minio-link/pkg/log"
	"github.com/google/uuid"
)

// Wrapper for YOURLS API (basic)
type YOURLSClient struct {
	logger    *log.Logger
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
	c.logger.Debug(fmt.Sprintf("(base) upload url: %s", u.String()))

	v := make(map[string]string)
	v["signature"] = c.signature
	v["action"] = "shorturl"
	v["format"] = "json"
	v["url"] = input
	v["title"] = defaultUploadTitle
	v["keyword"] = uuid.New().String()

	req, err := buildRequestWithContext(ctx, http.MethodPost, u.String(), createPostRequestBody(v))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}
	c.logger.Debug(
		fmt.Sprintf("url: %s, method: %s, body: %s", req.URL.String(), req.Method, req.Body),
	)

	res, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	c.logger.Debug(fmt.Sprintf("response: %s (%d)", res.Status, res.StatusCode))

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

	c.logger.Debug(fmt.Sprintf("shortened url: %s", shortenRes.Shorturl))

	return shortenRes.Shorturl, nil
}

// ExpandURL expands a shortened URL via YOURLS
func (c *YOURLSClient) ExpandURL(ctx context.Context, input string) (string, error) {
	_, err := checkURL(input)
	if err != nil {
		return "", fmt.Errorf("invalid input url: %w", err)
	}

	u, err := checkURL(fmt.Sprintf("%s/%s", c.baseURL, defaultAPIEndpoint))
	if err != nil {
		return "", fmt.Errorf("invalid base url: %w", err)
	}
	c.logger.Debug(fmt.Sprintf("(base) upload url: %s", u.String()))

	v := make(map[string]string)
	v["signature"] = c.signature
	v["action"] = "expand"
	v["format"] = "json"
	v["shorturl"] = input

	req, err := buildRequestWithContext(ctx, http.MethodPost, u.String(), createPostRequestBody(v))
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}
	c.logger.Debug(
		fmt.Sprintf("url: %s, method: %s, body: %s", req.URL.String(), req.Method, req.Body),
	)

	res, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	c.logger.Debug(fmt.Sprintf("response: %s (%d)", res.Status, res.StatusCode))

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

	c.logger.Debug(fmt.Sprintf("expanded url: %s", expandRes.Longurl))

	return expandRes.Longurl, nil
}

// Gets the shortened and saved urls of this program via stats endpoint of YOURLS api endpoint
func (c *YOURLSClient) GetSavedURLs(ctx context.Context, limit int) (map[string]string, error) {
	u, err := checkURL(fmt.Sprintf("%s/%s", c.baseURL, defaultAPIEndpoint))
	if err != nil {
		return nil, fmt.Errorf("invalid base url: %w", err)
	}
	c.logger.Debug(fmt.Sprintf("(base) list url: %s", u.String()))

	if limit < 1 {
		limit = defaultLinkLimit
		c.logger.Debug(
			fmt.Sprintf("limit is less than 1, set to default limit: %d", defaultLinkLimit),
		)
	}

	v := make(map[string]string)
	v["signature"] = c.signature
	v["action"] = "stats"
	v["format"] = "json"
	v["filter"] = "last"
	v["limit"] = strconv.Itoa(limit)

	req, err := buildRequestWithContext(ctx, http.MethodPost, u.String(), createPostRequestBody(v))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	c.logger.Debug(
		fmt.Sprintf("url: %s, method: %s, body: %s", req.URL.String(), req.Method, req.Body),
	)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	c.logger.Debug(fmt.Sprintf("response: %s (%d)", res.Status, res.StatusCode))

	if res.StatusCode != http.StatusOK {
		var errRes shortenURLErrorResponse
		if err := unmarshalResponseToJSON(res, &errRes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return nil, fmt.Errorf("failed to shorten url: %s", errRes.Message)
	}

	var data links

	if err := unmarshalResponseToJSON(res, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	urls := make(map[string]string)
	for _, link := range data.Links {
		if strings.Contains(link.Title, defaultUploadTitle) {
			urls[link.ShortURL] = link.URL
		}
	}

	c.logger.Debug(fmt.Sprintf("found %d urls", len(urls)))

	return urls, nil
}

// NewClient creates a new YOURLSClient
func NewClient(dir string, debug bool, cfg *environment.EnvConfig) *YOURLSClient {
	return &YOURLSClient{
		logger: log.NewLogger().
			WithDirectory(dir).
			WithName("yourls").
			WithDebug(debug).
			WithConsoleOutput(debug),
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
	defaultUploadTitle string = "Uploaded using minio-yourls-uploader by devusSs"
	defaultLinkLimit   int    = 20
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

type linkData struct {
	ShortURL  string `json:"shorturl"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Timestamp string `json:"timestamp"`
	IP        string `json:"ip"`
	Clicks    int    `json:"clicks"`
}

type links struct {
	Links map[string]linkData `json:"links"`
}
