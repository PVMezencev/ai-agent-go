package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"bytes"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// WebSearch implements the WebSearchInterface for real web operations
type WebSearch struct {
	config WebConfig
	client *http.Client
}

// NewWebSearch creates a new web search instance
func NewWebSearch(config WebConfig) (*WebSearch, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.UserAgent == "" {
		config.UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}

	return &WebSearch{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return errors.New("too many redirects")
				}
				return nil
			},
		},
	}, nil
}

// --- DuckDuckGo Search ---

// Search performs a real web search via DuckDuckGo.
// It first tries the Instant Answer JSON API (no key required),
// then falls back to HTML parsing of duckduckgo.com/html/.
func (ws *WebSearch) Search(ctx context.Context, query string, options SearchOptions) (*SearchResults, error) {
	if options.MaxResults == 0 {
		options.MaxResults = 10
	}
	if options.Timeout == 0 {
		options.Timeout = ws.config.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	// Try Instant Answer API first (JSON, no auth)
	results, err := ws.searchInstantAnswer(ctx, query, options)
	if err == nil && len(results) > 0 {
		return ws.buildResults(query, results, options.MaxResults), nil
	}

	// Fallback: parse HTML from duckduckgo.com/html/
	results, err = ws.searchHTML(ctx, query, options)
	if err != nil {
		return nil, fmt.Errorf("duckduckgo search failed: %w", err)
	}

	return ws.buildResults(query, results, options.MaxResults), nil
}

type ddgAPIResponse struct {
	Results         string            `json:"results"`
	Heading         string            `json:"heading"`
	AbstractText    string            `json:"abstract_text"`
	AbstractURL     string            `json:"abstract_url"`
	AbstractSource  string            `json:"abstract_source"`
	RelatedTopics   []ddgRelatedTopic `json:"RelatedTopics"`
}

type ddgRelatedTopic struct {
	Topic ddgTopic `json:"Topic"`
}

type ddgTopic struct {
	Text     string `json:"Text"`
	FirstURL string `json:"FirstURL"`
	Name     string `json:"Name"`
	Topics   []ddgTopic `json:"Topics"`
}

type searchHit struct {
	Title       string
	URL         string
	Description string
}

func (ws *WebSearch) searchInstantAnswer(ctx context.Context, query string, options SearchOptions) ([]searchHit, error) {
	encQuery := url.QueryEscape(query)
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_redirect=1&no_html=1&skip_disambig=1", encQuery)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	ws.setHeaders(req, options)

	resp, err := ws.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("duckduckgo api returned status %d", resp.StatusCode)
	}

	var answer ddgAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&answer); err != nil {
		return nil, err
	}

	var hits []searchHit

	// Abstract
	if answer.AbstractText != "" && answer.AbstractURL != "" {
		hits = append(hits, searchHit{
			Title:       answer.Heading,
			URL:         answer.AbstractURL,
			Description: answer.AbstractText,
		})
	}

	// Related topics (recursive)
	collectTopics(&answer.RelatedTopics, &hits)

	return hits, nil
}

func collectTopics(topics *[]ddgRelatedTopic, hits *[]searchHit) {
	if topics == nil {
		return
	}
	for _, rt := range *topics {
		t := rt.Topic
		if t.Text != "" && t.FirstURL != "" {
			title := t.Name
			if title == "" {
				title = truncate(t.Text, 80)
			}
			*hits = append(*hits, searchHit{
				Title:       title,
				URL:         t.FirstURL,
				Description: t.Text,
			})
		}
		// Recurse into nested topics
		if len(t.Topics) > 0 {
			nested := make([]ddgRelatedTopic, len(t.Topics))
			for i, nt := range t.Topics {
				nested[i] = ddgRelatedTopic{Topic: nt}
			}
			collectTopics(&nested, hits)
		}
	}
}

const ddgHTMLURL = "https://html.duckduckgo.com/html/?q="

func (ws *WebSearch) searchHTML(ctx context.Context, query string, options SearchOptions) ([]searchHit, error) {
	targetURL := ddgHTMLURL + url.QueryEscape(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	ws.setHeaders(req, options)

	resp, err := ws.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("duckduckgo html returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %w", err)
	}

	var hits []searchHit
	doc.Find(".result").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".result__title").Text()
		description := s.Find(".result__snippet").Text()

		link, _ := s.Find(".result__url a").Attr("href")
		if link == "" {
			link, _ = s.Find(".result__title a").Attr("href")
		}

		// DDG returns paths like /l/?uddg=https://example.com — extract the actual URL
		if strings.HasPrefix(link, "/l/") {
			if extracted := extractDDGURL(link); extracted != "" {
				link = extracted
			}
		}

		hits = append(hits, searchHit{
			Title:       strings.TrimSpace(title),
			URL:         link,
			Description: strings.TrimSpace(description),
		})
	})

	return hits, nil
}

// extractDDGURL pulls the actual URL from a DuckDuckGo redirect link
func extractDDGURL(link string) string {
	uddgPat := regexp.MustCompile(`uddg=([^&]+)`)
	if m := uddgPat.FindStringSubmatch(link); len(m) > 1 {
		u, err := url.QueryUnescape(m[1])
		if err == nil {
			return u
		}
	}

	dPat := regexp.MustCompile(`\bd=([^&]+)`)
	if m := dPat.FindStringSubmatch(link); len(m) > 1 {
		u, err := url.QueryUnescape(m[1])
		if err == nil {
			return u
		}
	}

	return ""
}

func (ws *WebSearch) buildResults(query string, hits []searchHit, maxResults int) *SearchResults {
	results := make([]SearchResult, 0, min(len(hits), maxResults))
	for i, h := range hits {
		if len(results) >= maxResults {
			break
		}
		results = append(results, SearchResult{
			Title:       h.Title,
			URL:         h.URL,
			Description: h.Description,
			Snippet:     h.Description,
			Source:      extractDomain(h.URL),
			LastUpdated: time.Now().Add(-time.Duration(i) * time.Hour),
		})
	}
	return &SearchResults{
		Query:     query,
		Results:   results,
		Total:     len(results),
		Processed: time.Now(),
	}
}

func (ws *WebSearch) setHeaders(req *http.Request, options SearchOptions) {
	ua := options.UserAgent
	if ua == "" {
		ua = ws.config.UserAgent
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")

	for k, v := range options.Headers {
		req.Header.Set(k, v)
	}
}

// SearchWithBrowser performs a web search (same as Search, with browser-like headers)
func (ws *WebSearch) SearchWithBrowser(ctx context.Context, query string, options SearchOptions) (*SearchResults, error) {
	if options.UserAgent == "" {
		options.UserAgent = ws.config.UserAgent
	}
	if options.Headers == nil {
		options.Headers = make(map[string]string)
	}
	options.Headers["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"
	options.Headers["Accept-Language"] = "en-US,en;q=0.5"
	options.Headers["Accept-Encoding"] = "gzip, deflate"
	options.Headers["Connection"] = "keep-alive"
	options.Headers["Upgrade-Insecure-Requests"] = "1"

	return ws.Search(ctx, query, options)
}

// --- Scraping ---

// Scrape fetches a real URL, parses the HTML, and extracts text content
func (ws *WebSearch) Scrape(ctx context.Context, scrapeURL string, options ScrapeOptions) (*ScrapeResult, error) {
	if options.Timeout == 0 {
		options.Timeout = ws.config.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, scrapeURL, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %s: %w", scrapeURL, err)
	}

	ua := options.UserAgent
	if ua == "" {
		ua = ws.config.UserAgent
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	for k, v := range options.Headers {
		req.Header.Set(k, v)
	}

	resp, err := ws.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", scrapeURL, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	result := &ScrapeResult{
		URL:        resp.Request.URL.String(),
		StatusCode: resp.StatusCode,
		Processed:  time.Now(),
		Meta:       make(map[string]string),
	}

	if resp.StatusCode >= 400 {
		result.Content = fmt.Sprintf("Server returned status %d", resp.StatusCode)
		return result, nil
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml") {
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
		if err != nil {
			result.Content = string(bodyBytes[:min(len(bodyBytes), 5000)])
			return result, nil
		}

		// Title
		result.Title = strings.TrimSpace(doc.Find("title").Text())

		// Meta tags
		doc.Find("meta").Each(func(i int, s *goquery.Selection) {
			name, _ := s.Attr("name")
			content, _ := s.Attr("content")
			if name != "" {
				result.Meta[name] = content
			}
			prop, _ := s.Attr("property")
			if prop != "" {
				result.Meta[prop] = content
			}
		})

		// Remove script/style, then get text
		doc.Find("script, style, noscript").Remove()

		// Get main content if possible
		var contentSel *goquery.Selection
		for _, sel := range []string{"article", "main", ".post", ".content", "#content"} {
			contentSel = doc.Find(sel)
			if contentSel.Length() > 0 {
				break
			}
		}

		var rawText string
		if contentSel != nil && contentSel.Length() > 0 {
			rawText = contentSel.Text()
		} else {
			rawText = doc.Find("body").Text()
		}

		result.Content = cleanText(rawText)

		// Extract links
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists && href != "" && !strings.HasPrefix(href, "javascript:") {
				result.Links = append(result.Links, href)
			}
		})

		// Extract images
		doc.Find("img").Each(func(i int, s *goquery.Selection) {
			src, exists := s.Attr("src")
			if exists && src != "" {
				result.Images = append(result.Images, src)
			}
		})

		// Limit counts
		if len(result.Links) > 100 {
			result.Links = result.Links[:100]
		}
		if len(result.Images) > 50 {
			result.Images = result.Images[:50]
		}

	} else {
		// Non-HTML content — return raw text (truncated)
		result.Content = string(bodyBytes[:min(len(bodyBytes), 10000)])
		result.Title = scrapeURL
	}

	return result, nil
}

// GetConfig returns the web configuration
func (ws *WebSearch) GetConfig() WebConfig {
	return ws.config
}

// SetConfig sets the web configuration
func (ws *WebSearch) SetConfig(config WebConfig) error {
	ws.config = config
	if config.Timeout > 0 {
		ws.client.Timeout = config.Timeout
	}
	return nil
}

// SetProxy sets the proxy for web operations
func (ws *WebSearch) SetProxy(proxy string) error {
	if proxy == "" {
		ws.client.Transport = nil
		return nil
	}

	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	ws.client.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	return nil
}

// --- Helpers ---

func extractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Host
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return strings.TrimSpace(s[:n]) + "…"
}

func cleanText(s string) string {
	lines := strings.Split(s, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, "\n")
}

// Ensure math package is used
var _ = math.Max
