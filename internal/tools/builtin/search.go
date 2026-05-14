package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shoucheng/my-first-agent/internal/tools"
)

type OmniSearch struct {
	client *http.Client
}

func NewOmniSearch() *OmniSearch {
	return &OmniSearch{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (t *OmniSearch) Name() string { return "omni_search" }

func (t *OmniSearch) Description() string {
	return tools.NormalizeDescription(
		"Unified search tool that performs comprehensive searches across web, images, and academic sources.",
		"When to use: comprehensive information gathering, current events, academic papers, image search.",
		"Best practices: use clear, specific queries with 3-7 keywords for optimal results.",
	)
}

func (t *OmniSearch) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"brief": tools.StringProperty("Brief description of search intent"),
		"search_type": map[string]any{
			"type":        "string",
			"description": "Type of search: info (web), image (images), api (documentation), news (web), tool (web+api), data/research (web+scholar+api)",
			"enum":        []any{"info", "image", "api", "news", "tool", "data", "research"},
		},
		"queries": map[string]any{
			"type":        "array",
			"description": "Search queries (1-3 queries max)",
			"items":       map[string]any{"type": "string"},
			"minItems":    1,
			"maxItems":    3,
		},
		"date_range": map[string]any{
			"type":        "string",
			"description": "Time range filter for search results",
			"enum":        []any{"all", "past_hour", "past_day", "past_week", "past_month", "past_year"},
		},
	}, "brief", "search_type", "queries")
}

func (t *OmniSearch) Execute(ctx context.Context, args map[string]any) (string, error) {
	queriesRaw, ok := args["queries"].([]any)
	if !ok || len(queriesRaw) == 0 {
		return "", fmt.Errorf("queries is required")
	}

	var results []string
	for _, q := range queriesRaw {
		query, _ := q.(string)
		if query == "" {
			continue
		}
		text, err := t.searchDuckDuckGo(ctx, query)
		if err != nil {
			results = append(results, fmt.Sprintf("Query %q: error: %v", query, err))
			continue
		}
		results = append(results, fmt.Sprintf("Query %q:\n%s", query, text))
	}

	if len(results) == 0 {
		return "No search results found.", nil
	}
	return strings.Join(results, "\n\n---\n\n"), nil
}

func (t *OmniSearch) searchDuckDuckGo(ctx context.Context, query string) (string, error) {
	endpoint := "https://api.duckduckgo.com/?format=json&no_redirect=1&no_html=1&q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("search request failed: %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}

	var payload struct {
		AbstractText  string `json:"AbstractText"`
		AbstractURL   string `json:"AbstractURL"`
		RelatedTopics []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
			Topics   []struct {
				Text     string `json:"Text"`
				FirstURL string `json:"FirstURL"`
			} `json:"Topics"`
		} `json:"RelatedTopics"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}

	var lines []string
	if payload.AbstractText != "" {
		if payload.AbstractURL != "" {
			lines = append(lines, fmt.Sprintf("Summary: %s (%s)", payload.AbstractText, payload.AbstractURL))
		} else {
			lines = append(lines, "Summary: "+payload.AbstractText)
		}
	}
	for _, topic := range payload.RelatedTopics {
		if len(lines) >= 8 {
			break
		}
		if topic.Text != "" {
			if topic.FirstURL != "" {
				lines = append(lines, fmt.Sprintf("- %s (%s)", topic.Text, topic.FirstURL))
			} else {
				lines = append(lines, "- "+topic.Text)
			}
		}
		for _, sub := range topic.Topics {
			if len(lines) >= 8 {
				break
			}
			if sub.Text != "" {
				if sub.FirstURL != "" {
					lines = append(lines, fmt.Sprintf("  - %s (%s)", sub.Text, sub.FirstURL))
				} else {
					lines = append(lines, "  - "+sub.Text)
				}
			}
		}
	}

	if len(lines) == 0 {
		return "No concise results found.", nil
	}
	return strings.Join(lines, "\n"), nil
}
