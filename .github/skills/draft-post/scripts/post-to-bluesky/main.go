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
	"unicode/utf8"

	"github.com/joho/godotenv"
)

const (
	bskyAPI = "https://bsky.social/xrpc"
)

// Session holds the auth session
type Session struct {
	DID       string `json:"did"`
	Handle    string `json:"handle"`
	AccessJwt string `json:"accessJwt"`
}

// PostRecord is the AT Protocol post record
type PostRecord struct {
	Type      string      `json:"$type"`
	Text      string      `json:"text"`
	CreatedAt string      `json:"createdAt"`
	Facets    []Facet     `json:"facets,omitempty"`
	Embed     interface{} `json:"embed,omitempty"`
}

// Facet represents a rich text annotation (links, mentions, etc.)
type Facet struct {
	Index    FacetIndex     `json:"index"`
	Features []FacetFeature `json:"features"`
}

type FacetIndex struct {
	ByteStart int `json:"byteStart"`
	ByteEnd   int `json:"byteEnd"`
}

type FacetFeature struct {
	Type string `json:"$type"`
	URI  string `json:"uri,omitempty"`
}

// ExternalEmbed is a link card embed
type ExternalEmbed struct {
	Type     string        `json:"$type"`
	External ExternalCard  `json:"external"`
}

type ExternalCard struct {
	URI         string `json:"uri"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func main() {
	_ = godotenv.Load()

	handle := os.Getenv("BSKY_HANDLE")
	appPassword := os.Getenv("BSKY_APP_PASSWORD")

	if handle == "" || appPassword == "" {
		fmt.Fprintln(os.Stderr, "Error: Set BSKY_HANDLE and BSKY_APP_PASSWORD in .env")
		fmt.Fprintln(os.Stderr, "Get an app password at: https://bsky.app/settings/app-passwords")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  go run main.go publish <markdown-file>   Post to Bluesky with link card")
		fmt.Fprintln(os.Stderr, "  go run main.go draft <markdown-file>     Show what would be posted (dry run)")
		fmt.Fprintln(os.Stderr, "  go run main.go me                        Test authentication")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Env vars: BSKY_HANDLE, BSKY_APP_PASSWORD")
		fmt.Fprintln(os.Stderr, "Optional: BSKY_DEVTO_URL (full URL of the dev.to article to link)")
		os.Exit(1)
	}

	command := os.Args[1]

	// Authenticate
	session, err := createSession(handle, appPassword)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Auth error: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "me":
		fmt.Printf("Authenticated as: @%s (DID: %s)\n", session.Handle, session.DID)
	case "draft":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: go run main.go draft <markdown-file>")
			os.Exit(1)
		}
		title, _, _ := parseMarkdownFile(os.Args[2])
		text := buildPostText(title)
		fmt.Println("=== DRY RUN (would post) ===")
		fmt.Printf("Text (%d chars):\n%s\n", len(text), text)
		fmt.Println("Draft ID: dry-run")
	case "publish":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: go run main.go publish <markdown-file>")
			os.Exit(1)
		}
		title, tags, _ := parseMarkdownFile(os.Args[2])
		devtoURL := os.Getenv("BSKY_DEVTO_URL")
		postText := buildPostText(title)

		// Add hashtags from tags, respecting 300 grapheme limit
		hashtagLine := ""
		for _, t := range tags {
			t = strings.TrimSpace(t)
			if t != "" {
				candidate := hashtagLine + "#" + t + " "
				if utf8.RuneCountInString(postText+"\n\n"+strings.TrimSpace(candidate)) > 300 {
					break // stop adding hashtags if it would exceed limit
				}
				hashtagLine = candidate
			}
		}
		if hashtagLine != "" {
			postText += "\n\n" + strings.TrimSpace(hashtagLine)
		}

		createPost(session, postText, title, devtoURL)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func buildPostText(title string) string {
	base := fmt.Sprintf("New post: %s\n\nRead the full article below 👇", title)
	// Bluesky has 300 grapheme limit; truncate title if needed
	if utf8.RuneCountInString(base) > 250 {
		// Leave room for hashtags (~50 chars)
		maxTitleRunes := 250 - utf8.RuneCountInString("New post: \n\nRead the full article below 👇")
		runes := []rune(title)
		if len(runes) > maxTitleRunes {
			title = string(runes[:maxTitleRunes-1]) + "…"
		}
		base = fmt.Sprintf("New post: %s\n\nRead the full article below 👇", title)
	}
	return base
}

// createSession authenticates with Bluesky using handle + app password
func createSession(handle, appPassword string) (*Session, error) {
	payload := map[string]string{
		"identifier": handle,
		"password":   appPassword,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(bskyAPI+"/com.atproto.server.createSession", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("auth failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var session Session
	if err := json.Unmarshal(respBody, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// createPost creates a post on Bluesky with optional link card
func createPost(session *Session, text, title, linkURL string) {
	record := PostRecord{
		Type:      "app.bsky.feed.post",
		Text:      text,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Add link card embed if URL provided
	if linkURL != "" {
		record.Embed = ExternalEmbed{
			Type: "app.bsky.embed.external",
			External: ExternalCard{
				URI:         linkURL,
				Title:       title,
				Description: "Read the full article on dev.to",
			},
		}

		// Add link facet so the URL is clickable in the text
		// Use a short label instead of the full URL to stay under 300 graphemes
		linkLabel := "Read here →"
		record.Text = text + "\n\n" + linkLabel

		byteStart := len([]byte(text + "\n\n"))
		byteEnd := byteStart + len([]byte(linkLabel))
		record.Facets = []Facet{
			{
				Index: FacetIndex{ByteStart: byteStart, ByteEnd: byteEnd},
				Features: []FacetFeature{
					{Type: "app.bsky.richtext.facet#link", URI: linkURL},
				},
			},
		}
	}

	payload := map[string]interface{}{
		"repo":       session.DID,
		"collection": "app.bsky.feed.post",
		"record":     record,
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", bskyAPI+"/com.atproto.repo.createRecord", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+session.AccessJwt)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Post error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "Post failed (%d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var result struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}
	json.Unmarshal(respBody, &result)
	fmt.Printf("URL: https://bsky.app/profile/%s/post/%s\n", session.Handle, extractPostID(result.URI))
}

// extractPostID gets the post ID from an AT URI like at://did:plc:xxx/app.bsky.feed.post/abc123
func extractPostID(atURI string) string {
	parts := strings.Split(atURI, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return atURI
}

// parseMarkdownFile reads a markdown file and extracts YAML frontmatter
func parseMarkdownFile(path string) (title string, tags []string, body string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	content := string(data)
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content[3:], "---", 2)
		if len(parts) == 2 {
			frontmatter := parts[0]
			body = strings.TrimSpace(parts[1])

			for _, line := range strings.Split(frontmatter, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "title:") {
					title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
				}
				if strings.HasPrefix(line, "tags:") {
					tagStr := strings.TrimSpace(strings.TrimPrefix(line, "tags:"))
					for _, t := range strings.Split(tagStr, ",") {
						tags = append(tags, strings.TrimSpace(t))
					}
				}
			}
		}
	}
	return
}
