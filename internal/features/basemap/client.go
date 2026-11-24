package basemap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultStyle    = "voyager"
	defaultTemplate = "https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png"
	apiKeyParam     = "key"
	maxZoom         = 20
	subdomainCount  = 4
	maxTileBytes    = 2 * 1024 * 1024
	requestTimeout  = 5 * time.Second
	retinaSuffix    = "@2x"
)

// Kesalahan yang dapat dipetakan handler ke status HTTP.
var (
	ErrNotConfigured    = errors.New("basemap api key not configured")
	ErrUnsupportedStyle = errors.New("unsupported basemap style")
	ErrInvalidTile      = errors.New("invalid tile coordinates")
)

// Client mengambil tile raster dari penyedia basemap.
type Client struct {
	apiKey string
	styles map[string]string
	http   *http.Client
}

// NewClient membuat client dengan timeout 5 detik dan template style voyager.
func NewClient(apiKey, template string) *Client {
	if template == "" {
		template = defaultTemplate
	}
	return &Client{
		apiKey: apiKey,
		styles: map[string]string{defaultStyle: template},
		http:   &http.Client{Timeout: requestTimeout},
	}
}

// Fetch mengambil satu tile untuk koordinat, mode retina, dan style tertentu.
func (c *Client) Fetch(ctx context.Context, z, x, y int, retina bool, style string) ([]byte, error) {
	if c.apiKey == "" {
		return nil, ErrNotConfigured
	}
	rawURL, err := c.tileURL(z, x, y, retina, style)
	if err != nil {
		return nil, err
	}
	return c.download(ctx, rawURL)
}

func (c *Client) tileURL(z, x, y int, retina bool, style string) (string, error) {
	template, ok := c.styles[style]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrUnsupportedStyle, style)
	}
	if err := validateTile(z, x, y); err != nil {
		return "", err
	}
	suffix := ""
	if retina {
		suffix = retinaSuffix
	}
	replacer := strings.NewReplacer(
		"{s}", string(rune('a'+(x+y)%subdomainCount)),
		"{z}", strconv.Itoa(z),
		"{x}", strconv.Itoa(x),
		"{y}", strconv.Itoa(y),
		"{r}", suffix,
	)
	return replacer.Replace(template), nil
}

func validateTile(z, x, y int) error {
	if z < 0 || z > maxZoom {
		return fmt.Errorf("%w: zoom %d", ErrInvalidTile, z)
	}
	limit := 1 << z
	if x < 0 || x >= limit || y < 0 || y >= limit {
		return fmt.Errorf("%w: z=%d x=%d y=%d", ErrInvalidTile, z, x, y)
	}
	return nil
}

func (c *Client) download(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := c.buildRequest(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("basemap request: %w", sanitizeRequestError(err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("basemap upstream status %d", resp.StatusCode)
	}
	return readLimited(resp.Body)
}

func (c *Client) buildRequest(ctx context.Context, rawURL string) (*http.Request, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse basemap url: %w", err)
	}
	q := u.Query()
	q.Set(apiKeyParam, c.apiKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build basemap request: %w", sanitizeRequestError(err))
	}
	return req, nil
}

// sanitizeRequestError membuang seluruh URL dari rantai error transport agar key tidak bocor.
func sanitizeRequestError(err error) error {
	for {
		var urlErr *url.Error
		if !errors.As(err, &urlErr) {
			return err
		}
		err = urlErr.Err
		if err == nil {
			return errors.New("basemap request failed")
		}
	}
}

func readLimited(body io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(body, maxTileBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read basemap tile: %w", err)
	}
	if len(data) > maxTileBytes {
		return nil, fmt.Errorf("basemap tile exceeds %d bytes", maxTileBytes)
	}
	return data, nil
}
