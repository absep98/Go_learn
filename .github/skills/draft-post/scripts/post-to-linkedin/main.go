package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	authURL     = "https://www.linkedin.com/oauth/v2/authorization"
	tokenURL    = "https://www.linkedin.com/oauth/v2/accessToken"
	userInfoURL = "https://api.linkedin.com/v2/userinfo"
	postsURL    = "https://api.linkedin.com/rest/posts"
	redirectURI = "http://localhost:5000/auth/linkedin/callback"
	apiVersion  = "202604"
)

func main() {
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  go run main.go login                    One-time OAuth login (saves token to .env)")
		fmt.Fprintln(os.Stderr, "  go run main.go me                        Test authentication")
		fmt.Fprintln(os.Stderr, "  go run main.go publish <markdown-file>   Post to LinkedIn")
		fmt.Fprintln(os.Stderr, "  go run main.go draft <markdown-file>     Show what would be posted (dry run)")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "login":
		doLogin()
	case "me":
		token := requireToken()
		doMe(token)
	case "draft":
		requireFile()
		title, tags, _ := parseMarkdownFile(os.Args[2])
		text := buildPostText(title, tags)
		devtoURL := getDevtoURL()
		fmt.Println("=== DRY RUN (would post) ===")
		fmt.Printf("Text (%d chars):\n%s\n", len(text), text)
		if devtoURL != "" {
			fmt.Printf("Link card: %s\n", devtoURL)
		}
		fmt.Println("Draft ID: dry-run")
	case "publish":
		requireFile()
		token := requireToken()
		// Use organization URN if set, otherwise fall back to person URN
		authorURN := os.Getenv("LINKEDIN_ORG_URN")
		if authorURN == "" {
			authorURN = os.Getenv("LINKEDIN_PERSON_URN")
			if authorURN == "" {
				authorURN = fetchAndSavePersonURN(token)
			}
		}
		title, tags, _ := parseMarkdownFile(os.Args[2])
		text := buildPostText(title, tags)
		devtoURL := getDevtoURL()
		doPublish(token, authorURN, text, title, devtoURL)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func requireToken() string {
	token := os.Getenv("LINKEDIN_ACCESS_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: LINKEDIN_ACCESS_TOKEN not set. Run 'go run main.go login' first.")
		os.Exit(1)
	}
	return token
}

func requireFile() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: go run main.go %s <markdown-file>\n", os.Args[1])
		os.Exit(1)
	}
}

func getDevtoURL() string {
	u := os.Getenv("LINKEDIN_DEVTO_URL")
	if u == "" {
		u = os.Getenv("DEVTO_URL") // fallback for post-everywhere.ps1
	}
	return u
}

// ── OAuth Login Flow ─────────────────────────────────────────

func doLogin() {
	clientID := os.Getenv("LINKEDIN_CLIENT_ID")
	clientSecret := os.Getenv("LINKEDIN_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		fmt.Fprintln(os.Stderr, "Error: Set LINKEDIN_CLIENT_ID and LINKEDIN_CLIENT_SECRET in .env")
		os.Exit(1)
	}

	// Build authorization URL
	params := url.Values{
		"response_type": {"code"},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"scope":         {"openid profile w_member_social w_organization_social email"},
	}
	loginURL := authURL + "?" + params.Encode()

	// Channel to receive the auth code
	codeCh := make(chan string, 1)
	errCh := make(chan string, 1)

	// Start local server to capture the callback
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/linkedin/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		errMsg := r.URL.Query().Get("error")
		if errMsg != "" {
			errCh <- r.URL.Query().Get("error_description")
			fmt.Fprintf(w, "<h1>Error: %s</h1><p>%s</p>", errMsg, r.URL.Query().Get("error_description"))
			return
		}
		if code == "" {
			errCh <- "no code received"
			fmt.Fprint(w, "<h1>Error: No authorization code received</h1>")
			return
		}
		codeCh <- code
		fmt.Fprint(w, "<h1>Success! You can close this tab.</h1><p>Go back to your terminal.</p>")
	})

	server := &http.Server{Addr: ":5000", Handler: mux}
	go func() { _ = server.ListenAndServe() }()

	fmt.Println("Opening browser for LinkedIn login...")
	fmt.Printf("If browser doesn't open, visit:\n%s\n\n", loginURL)
	openBrowser(loginURL)

	// Wait for callback
	var code string
	select {
	case code = <-codeCh:
		fmt.Println("Authorization code received!")
	case e := <-errCh:
		fmt.Fprintf(os.Stderr, "Login failed: %s\n", e)
		os.Exit(1)
	case <-time.After(2 * time.Minute):
		fmt.Fprintln(os.Stderr, "Timeout: No login within 2 minutes")
		os.Exit(1)
	}
	server.Close()

	// Exchange code for access token
	tokenData := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}

	resp, err := http.PostForm(tokenURL, tokenData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Token exchange error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "Token exchange failed (%d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	json.Unmarshal(body, &tokenResp)

	// Get person URN
	personURN := fetchPersonURN(tokenResp.AccessToken)

	// Save to .env
	saveToEnv(tokenResp.AccessToken, personURN)

	fmt.Printf("\nLogin successful!\n")
	fmt.Printf("Access token saved to .env (expires in %d days)\n", tokenResp.ExpiresIn/86400)
	fmt.Printf("Person URN: %s\n", personURN)
}

func fetchPersonURN(token string) string {
	req, _ := http.NewRequest("GET", userInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "UserInfo error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info struct {
		Sub  string `json:"sub"`
		Name string `json:"name"`
	}
	json.Unmarshal(body, &info)

	if info.Sub == "" {
		fmt.Fprintf(os.Stderr, "Failed to get user info: %s\n", string(body))
		os.Exit(1)
	}

	fmt.Printf("Logged in as: %s\n", info.Name)
	return "urn:li:person:" + info.Sub
}

func fetchAndSavePersonURN(token string) string {
	urn := fetchPersonURN(token)
	// Update .env with person URN
	data, err := os.ReadFile(".env")
	if err == nil {
		content := string(data)
		if strings.Contains(content, "LINKEDIN_PERSON_URN=") {
			lines := strings.Split(content, "\n")
			for i, line := range lines {
				if strings.HasPrefix(line, "LINKEDIN_PERSON_URN=") {
					lines[i] = "LINKEDIN_PERSON_URN=" + urn
				}
			}
			os.WriteFile(".env", []byte(strings.Join(lines, "\n")), 0600)
		} else {
			f, _ := os.OpenFile(".env", os.O_APPEND|os.O_WRONLY, 0600)
			f.WriteString("\nLINKEDIN_PERSON_URN=" + urn + "\n")
			f.Close()
		}
	}
	return urn
}

func saveToEnv(token, personURN string) {
	data, err := os.ReadFile(".env")
	if err != nil {
		// Create new .env
		content := fmt.Sprintf("LINKEDIN_CLIENT_ID=%s\nLINKEDIN_CLIENT_SECRET=%s\nLINKEDIN_ACCESS_TOKEN=%s\nLINKEDIN_PERSON_URN=%s\nLINKEDIN_DEVTO_URL=\n",
			os.Getenv("LINKEDIN_CLIENT_ID"), os.Getenv("LINKEDIN_CLIENT_SECRET"), token, personURN)
		os.WriteFile(".env", []byte(content), 0600)
		return
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	tokenSet, urnSet := false, false

	for i, line := range lines {
		if strings.HasPrefix(line, "LINKEDIN_ACCESS_TOKEN=") {
			lines[i] = "LINKEDIN_ACCESS_TOKEN=" + token
			tokenSet = true
		}
		if strings.HasPrefix(line, "LINKEDIN_PERSON_URN=") {
			lines[i] = "LINKEDIN_PERSON_URN=" + personURN
			urnSet = true
		}
	}

	if !tokenSet {
		lines = append(lines, "LINKEDIN_ACCESS_TOKEN="+token)
	}
	if !urnSet {
		lines = append(lines, "LINKEDIN_PERSON_URN="+personURN)
	}

	os.WriteFile(".env", []byte(strings.Join(lines, "\n")), 0600)
}

// ── Me Command ───────────────────────────────────────────────

func doMe(token string) {
	req, _ := http.NewRequest("GET", userInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info struct {
		Sub   string `json:"sub"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	json.Unmarshal(body, &info)
	fmt.Printf("Authenticated as: %s (sub: %s, email: %s)\n", info.Name, info.Sub, info.Email)
}

// ── Publish Command ──────────────────────────────────────────

func doPublish(token, personURN, text, title, devtoURL string) {
	post := map[string]interface{}{
		"author":         personURN,
		"commentary":     text,
		"visibility":     "PUBLIC",
		"lifecycleState": "PUBLISHED",
		"distribution": map[string]interface{}{
			"feedDistribution":               "MAIN_FEED",
			"targetEntities":                 []interface{}{},
			"thirdPartyDistributionChannels": []interface{}{},
		},
		"isReshareDisabledByAuthor": false,
	}

	// Add article link card if URL provided
	if devtoURL != "" {
		post["content"] = map[string]interface{}{
			"article": map[string]interface{}{
				"source":      devtoURL,
				"title":       title,
				"description": "Read the full article on dev.to",
			},
		}
	}

	body, _ := json.Marshal(post)

	req, _ := http.NewRequest("POST", postsURL, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("LinkedIn-Version", apiVersion)
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Post error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "Post failed (%d): %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	// Extract post URN from response header
	postURN := resp.Header.Get("X-Restli-Id")
	if postURN == "" {
		postURN = resp.Header.Get("x-restli-id")
	}

	fmt.Println("Posted to LinkedIn!")
	if postURN != "" {
		fmt.Printf("ID: %s\n", postURN)
	}
	fmt.Printf("URL: https://www.linkedin.com/feed/\n")
}

// ── Helpers ──────────────────────────────────────────────────

func buildPostText(title string, tags []string) string {
	text := fmt.Sprintf("%s\n\nNew post from my \"Learning Go in Public\" series. Read the full article below 👇", title)

	if len(tags) > 0 {
		hashtags := ""
		for _, t := range tags {
			t = strings.TrimSpace(t)
			if t != "" {
				hashtags += "#" + t + " "
			}
		}
		if hashtags != "" {
			text += "\n\n" + strings.TrimSpace(hashtags)
		}
	}
	return text
}

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

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}
