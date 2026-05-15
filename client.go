package jackett

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Muxcore-Media/core/pkg/contracts"
)

type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type torznabFeed struct {
	XMLName xml.Name       `xml:"rss"`
	Channel torznabChannel `xml:"channel"`
}

type torznabChannel struct {
	Items []torznabItem `xml:"item"`
}

type torznabItem struct {
	Title       string        `xml:"title"`
	GUID        string        `xml:"guid"`
	Link        string        `xml:"link"`
	Description string        `xml:"description"`
	Attrs       []torznabAttr `xml:"http://torznab.com/schemas/2015/feed attr"`
}

type torznabAttr struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

func (c *Client) Search(ctx context.Context, query contracts.SearchQuery) ([]contracts.IndexerResult, error) {
	params := url.Values{
		"t":      {"search"},
		"q":      {query.Query},
		"apikey": {c.apiKey},
	}
	if query.Limit > 0 {
		params.Set("limit", strconv.Itoa(query.Limit))
	}
	if len(query.Categories) > 0 {
		// Map common media types to Newznab categories
		cats := mapCategories(query.Categories)
		if cats != "" {
			params.Set("cat", cats)
		}
	}

	u := c.baseURL + "/api/v2.0/indexers/all/results/torznab?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("search request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed (status %d): %s", resp.StatusCode, string(body))
	}

	var feed torznabFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("decode results: %w", err)
	}

	results := make([]contracts.IndexerResult, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		r := contracts.IndexerResult{
			Title:      item.Title,
			GUID:       item.GUID,
			Link:       item.Link,
			Categories: []string{},
			Source:     "jackett",
		}
		for _, attr := range item.Attrs {
			switch attr.Name {
			case "magneturl":
				r.MagnetURI = attr.Value
			case "size":
				r.Size, _ = strconv.ParseInt(attr.Value, 10, 64)
			case "seeders":
				r.Seeders, _ = strconv.Atoi(attr.Value)
			case "peers":
				r.Leechers, _ = strconv.Atoi(attr.Value)
			case "publishdate":
				r.PublishDate = attr.Value
			case "category":
				r.Categories = append(r.Categories, attr.Value)
			default:
				if r.Extra == nil {
					r.Extra = make(map[string]string)
				}
				r.Extra[attr.Name] = attr.Value
			}
		}
		results = append(results, r)
	}
	return results, nil
}

func (c *Client) Name() string {
	return "jackett"
}

func (c *Client) Capabilities(ctx context.Context) ([]string, error) {
	return []string{"search", "torznab", "newznab"}, nil
}

// mapCategories maps common media type categories to Newznab category IDs.
func mapCategories(categories []string) string {
	mapping := map[string]string{
		"movie":     "2000",
		"tv":        "5000",
		"music":     "3000",
		"book":      "7000",
		"anime":     "5070",
		"manga":     "7070",
		"audiobook": "3030",
	}
	var ids []string
	for _, cat := range categories {
		if id, ok := mapping[strings.ToLower(cat)]; ok {
			ids = append(ids, id)
		}
	}
	return strings.Join(ids, ",")
}
