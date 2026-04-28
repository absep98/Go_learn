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

// Reddit OAuth2 endpoints
const (
	redditTokenURL = "https://www.reddit.com/api/v1/access_token"
	redditSubmitURL = "https://oauth.reddit.com/api/submit"
)

func main() {
	_ = godotenv.Load()

	clientID := os.Getenv("REDDIT_CLIENT_ID")
	clientSecret := os.Getenv("REDDIT_CLIENT_SECRET")
	username := os.Getenv("REDDIT_USERNAME")
	password := os.Getenv("REDDIT_PASSWORD")
	subreddit := os.Getenv("REDDIT_SUBREDDIT") // e.g. "golang"

	if clientID == "" || clientSecret == "" || username == "" || password == "" {
		fmt.Fprintln(os.Stderr, "Error: Set REDDIT_CLIENT_ID, REDDIT_CLIENT_SECRET, REDDIT_USERNAME, REDDIT_PASSWORD in .env")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  go run main.go publish <markdown-file>   Post link to subreddit")
		fmt.Fprintln(os.Stderr, "  go run main.go draft <markdown-file>     Show what would be posted (dry run)")
		fmt.Fprintln(os.Stderr, "  go run main.go me                        Test authentication")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Env vars: REDDIT_CLIENT_ID, REDDIT_CLIENT_SECRET, REDDIT_USERNAME, REDDIT_PASSWORD, REDDIT_SUBREDDIT")
		fmt.Fprintln(os.Stderr, "REDDIT_DEVTO_URL_PREFIX should be set to your dev.to profile URL prefix for link posts")
		os.Exit(1)
	}

	command := os.Args[1]

	// Get OAuth2 token
	token, err := getAccessToken(clientID, clientSecret, username, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Auth error: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "me":
		me(token)
	case "draft":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: go run main.go draft <markdown-file>")
			os.Exit(1)
		}
		title, _, _ := parseMarkdownFile(os.Args[2])
		devtoURL := os.Getenv("REDDIT_DEVTO_URL_PREFIX")
		fmt.Println("=== DRY RUN (would post) ===")
		fmt.Printf("Subreddit: r/%s\n", subreddit)
		fmt.Printf("Title: %s\n", title)
		fmt.Printf("URL: %s (set REDDIT_DEVTO_URL_PREFIX)\n", devtoURL)
		fmt.Println("Draft ID: dry-run")
	case "publish":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: go run main.go publish <markdown-file>")
			os.Exit(1)
		}
		if subreddit == "" {
			subreddit = "golang"
		}
		title, _, _ := parseMarkdownFile(os.Args[2])
		devtoURL := os.Getenv("REDDIT_DEVTO_URL_PREFIX")
		if devtoURL == "" {
			fmt.Fprintln(os.Stderr, "Error: REDDIT_DEVTO_URL_PREFIX not set (your dev.to article URL)")
			os.Exit(1)
		}
		submitLink(token, subreddit, title, devtoURL)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

// getAccessToken uses Reddit's "script" app OAuth2 flow (password grant)
func getAccessToken(clientID, clientSecret, username, password string) (string, error) {
	data := fmt.Sprintf("grant_type=password&username=%s&password=%s", username, password)
	req, err := http.NewRequest("POST", redditTokenURL, strings.NewReader(data))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "GoLearnBot/1.0 (by /u/"+username+")")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("token request failed (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("auth error: %s", result.Error)
	}
	return result.AccessToken, nil
}

// me tests authentication by fetching user info
func me(token string) {
	req, _ := http.NewRequest("GET", "https://oauth.reddit.com/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "GoLearnBot/1.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}
	json.Unmarshal(body, &result)
	fmt.Printf("Authenticated as: /u/%s (ID: %s)\n", result.Name, result.ID)
}

// submitLink posts a link to a subreddit
func submitLink(token, subreddit, title, url string) {
	data := fmt.Sprintf("sr=%s&kind=link&title=%s&url=%s&resubmit=true",
		subreddit,
		strings.ReplaceAll(title, " ", "+"),
		strings.ReplaceAll(url, " ", "+"),
	)

	req, _ := http.NewRequest("POST", redditSubmitURL, strings.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "GoLearnBot/1.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Submit error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Success bool `json:"success"`
		JSON    struct {
			Data struct {
				URL string `json:"url"`
				ID  string `json:"id"`
			} `json:"data"`
			Errors [][]interface{} `json:"errors"`
		} `json:"json"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %s\n", string(body))
		os.Exit(1)
	}

	if len(result.JSON.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Reddit API error: %v\n", result.JSON.Errors)
		os.Exit(1)
	}

	fmt.Printf("URL: %s\n", result.JSON.Data.URL)
}

// parseMarkdownFile reads a markdown file and extracts YAML frontmatter
func parseMarkdownFile(path string) (title string, tags []string, body string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	content := string(data)
	// Extract frontmatter
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

// For link posts we need to use URL encoding, not bytes.Buffer
var _ = bytes.NewBuffer // suppress unused import (shared go.mod pattern)
