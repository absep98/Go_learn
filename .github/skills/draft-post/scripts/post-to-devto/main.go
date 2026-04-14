package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const devtoAPI = "https://dev.to/api/articles"

type Article struct {
	Title        string   `json:"title"`
	BodyMarkdown string   `json:"body_markdown"`
	Published    bool     `json:"published"`
	Tags         []string `json:"tags,omitempty"`
	Series       string   `json:"series,omitempty"`
}

type ArticleRequest struct {
	Article Article `json:"article"`
}

func main() {
	_ = godotenv.Load()

	apiKey := os.Getenv("DEVTO_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: DEVTO_API_KEY not set. Add it to .env or environment")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  go run main.go publish <markdown-file>      Post article (published)")
		fmt.Fprintln(os.Stderr, "  go run main.go draft <markdown-file>        Save as draft")
		fmt.Fprintln(os.Stderr, "  go run main.go publish-id <article-id>      Publish an existing draft")
		fmt.Fprintln(os.Stderr, "  go run main.go list                         List your articles")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Markdown file should have YAML frontmatter:")
		fmt.Fprintln(os.Stderr, "  ---")
		fmt.Fprintln(os.Stderr, "  title: My Article Title")
		fmt.Fprintln(os.Stderr, "  tags: go, buildinpublic, backend")
		fmt.Fprintln(os.Stderr, "  series: Learning Go in Public")
		fmt.Fprintln(os.Stderr, "  ---")
		fmt.Fprintln(os.Stderr, "  Article body here...")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "publish":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: provide a markdown file path")
			os.Exit(1)
		}
		err := postArticle(apiKey, os.Args[2], true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
			os.Exit(1)
		}
	case "draft":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: provide a markdown file path")
			os.Exit(1)
		}
		err := postArticle(apiKey, os.Args[2], false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
			os.Exit(1)
		}
	case "publish-id":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: provide an article ID")
			os.Exit(1)
		}
		err := publishDraft(apiKey, os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
			os.Exit(1)
		}
	case "list":
		err := listArticles(apiKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func postArticle(apiKey, filePath string, published bool) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	content := string(data)
	title, tags, series, body := parseFrontmatter(content)

	if title == "" {
		return fmt.Errorf("markdown file must have 'title' in frontmatter")
	}

	article := ArticleRequest{
		Article: Article{
			Title:        title,
			BodyMarkdown: body,
			Published:    published,
			Tags:         tags,
			Series:       series,
		},
	}

	jsonBody, err := json.Marshal(article)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest("POST", devtoAPI, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID  int    `json:"id"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil {
		status := "PUBLISHED"
		if !published {
			status = "DRAFT"
		}
		fmt.Printf("Article %s!\n", status)
		fmt.Printf("ID: %d\n", result.ID)
		fmt.Printf("URL: %s\n", result.URL)
	}

	return nil
}

func publishDraft(apiKey, articleID string) error {
	body := map[string]interface{}{
		"article": map[string]interface{}{
			"published": true,
		},
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest("PUT", devtoAPI+"/"+articleID, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil {
		fmt.Println("Article PUBLISHED!")
		fmt.Printf("URL: %s\n", result.URL)
	}

	return nil
}

func listArticles(apiKey string) error {
	req, err := http.NewRequest("GET", devtoAPI+"/me?per_page=10", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("api-key", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var articles []struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		URL         string `json:"url"`
		Published   bool   `json:"published"`
		PublishedAt string `json:"published_at"`
	}

	if err := json.Unmarshal(respBody, &articles); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	if len(articles) == 0 {
		fmt.Println("No articles yet. Start writing!")
		return nil
	}

	fmt.Printf("%-6s %-10s %s\n", "ID", "Status", "Title")
	fmt.Println(strings.Repeat("-", 60))
	for _, a := range articles {
		status := "draft"
		if a.Published {
			status = "published"
		}
		fmt.Printf("%-6d %-10s %s\n", a.ID, status, a.Title)
	}

	return nil
}

// parseFrontmatter extracts title, tags, series and body from markdown with YAML frontmatter
func parseFrontmatter(content string) (title string, tags []string, series string, body string) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return "", nil, "", content
	}

	// Find closing ---
	rest := content[3:]
	idx := strings.Index(rest, "---")
	if idx == -1 {
		return "", nil, "", content
	}

	frontmatter := rest[:idx]
	body = strings.TrimSpace(rest[idx+3:])

	for _, line := range strings.Split(frontmatter, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
		} else if strings.HasPrefix(line, "tags:") {
			tagStr := strings.TrimSpace(strings.TrimPrefix(line, "tags:"))
			for _, t := range strings.Split(tagStr, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tags = append(tags, t)
				}
			}
		} else if strings.HasPrefix(line, "series:") {
			series = strings.TrimSpace(strings.TrimPrefix(line, "series:"))
		}
	}

	return title, tags, series, body
}
