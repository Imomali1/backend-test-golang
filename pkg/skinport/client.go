package skinport

import (
	errs "backend-test-golang/pkg/errors"
	"backend-test-golang/pkg/ratelimiter"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/andybalholm/brotli"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultAppID    = 730
	defaultCurrency = "EUR"
)

type Client struct {
	clientID    string
	secretKey   string
	baseURL     string
	client      *http.Client
	rateLimiter *ratelimiter.RateLimiter
}

func NewClient(clientID, secretKey, baseURL string) (*Client, error) {
	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse skinport base url: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("invalid skinport url scheme")
	}

	if u.Host == "" {
		return nil, errors.New("missing skinport url host")
	}

	return &Client{
		clientID:    clientID,
		secretKey:   secretKey,
		baseURL:     baseURL,
		client:      &http.Client{Timeout: 10 * time.Second},
		rateLimiter: ratelimiter.New(8, 5*time.Minute), // skinport API's rate limits
	}, nil
}

type Item struct {
	MarketHashName string   `json:"market_hash_name"`
	Currency       string   `json:"currency"`
	SuggestedPrice *float64 `json:"suggested_price"`
	ItemPage       string   `json:"item_page"`
	MarketPage     string   `json:"market_page"`
	MinPrice       *float64 `json:"min_price"`
	MaxPrice       *float64 `json:"max_price"`
	MeanPrice      *float64 `json:"mean_price"`
	MedianPrice    *float64 `json:"median_price"`
	Quantity       int64    `json:"quantity"`
	CreatedAt      int64    `json:"created_at"`
	UpdatedAt      int64    `json:"updated_at"`
	Tradable       bool     `json:"tradable"`
}

func (c *Client) GetItems(ctx context.Context, tradable bool) ([]Item, error) {
	if !c.rateLimiter.Allow() {
		return nil, errs.NewRateLimitExceedErr(c.rateLimiter.RetryAfter())
	}

	u, _ := url.Parse(c.baseURL)
	u = u.JoinPath("/items")

	q := u.Query()
	q.Set("currency", defaultCurrency)
	q.Set("app_id", fmt.Sprintf("%d", defaultAppID))
	q.Set("tradable", fmt.Sprintf("%t", tradable))
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if c.clientID != "" && c.secretKey != "" {
		req.SetBasicAuth(c.clientID, c.secretKey)
	}

	req.Header.Set("Accept-Encoding", "br")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, c.tooManyRequestErr(resp.Header.Get("Retry-After"))
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("skinport API returned status %d: %s", resp.StatusCode, string(body))
	}

	var items []Item
	if err = json.NewDecoder(io.NopCloser(brotli.NewReader(resp.Body))).Decode(&items); err != nil {
		return nil, err
	}

	for i := range items {
		items[i].Tradable = tradable
	}

	return items, nil
}

func (c *Client) tooManyRequestErr(retryAfter string) error {
	if retryAfter == "" {
		c.rateLimiter.ForceFill()
		return errs.NewRateLimitExceedErr(c.rateLimiter.RetryAfter())
	}

	seconds, err := strconv.Atoi(retryAfter)
	if err != nil {
		log.Printf("could not parse retry after %s: %v", retryAfter, err)
		c.rateLimiter.ForceFill()
		return errs.NewRateLimitExceedErr(c.rateLimiter.RetryAfter())
	}

	until := time.Now().Add(time.Duration(seconds) * time.Second)
	c.rateLimiter.BlockUntil(until)
	return errs.NewRateLimitExceedErr(c.rateLimiter.RetryAfter())
}
