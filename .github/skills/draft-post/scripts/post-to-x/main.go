package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	tweetURLv2 = "https://api.twitter.com/2/tweets"
	tweetURLv1 = "https://api.twitter.com/1.1/statuses/update.json"
)

func main() {
	// Load .env from script directory or working directory
	_ = godotenv.Load()

	consumerKey := os.Getenv("X_CONSUMER_KEY")
	consumerSecret := os.Getenv("X_CONSUMER_SECRET")
	accessToken := os.Getenv("X_ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("X_ACCESS_TOKEN_SECRET")

	if consumerKey == "" || consumerSecret == "" || accessToken == "" || accessTokenSecret == "" {
		fmt.Fprintln(os.Stderr, "Error: missing credentials. Set X_CONSUMER_KEY, X_CONSUMER_SECRET, X_ACCESS_TOKEN, X_ACCESS_TOKEN_SECRET in .env or environment")
		os.Exit(1)
	}

	// Debug: show credential lengths to verify .env loaded correctly (no secrets printed)
	if os.Getenv("X_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[debug] Consumer Key: %d chars, starts with %s...\n", len(consumerKey), safePrefix(consumerKey, 4))
		fmt.Fprintf(os.Stderr, "[debug] Consumer Secret: %d chars\n", len(consumerSecret))
		fmt.Fprintf(os.Stderr, "[debug] Access Token: %d chars, starts with %s...\n", len(accessToken), safePrefix(accessToken, 4))
		fmt.Fprintf(os.Stderr, "[debug] Access Token Secret: %d chars\n", len(accessTokenSecret))
	}

	// Read tweet text from args or stdin
	tweetText := readTweetText()
	if tweetText == "" {
		fmt.Fprintln(os.Stderr, "Error: no tweet text provided. Pass as argument or pipe via stdin")
		fmt.Fprintln(os.Stderr, "Usage: go run main.go \"Your tweet text here\"")
		fmt.Fprintln(os.Stderr, "   or: echo \"Your tweet\" | go run main.go")
		os.Exit(1)
	}

	if len(tweetText) > 280 {
		fmt.Fprintf(os.Stderr, "Warning: tweet is %d characters (limit 280)\n", len(tweetText))
	}

	err := postTweet(tweetText, consumerKey, consumerSecret, accessToken, accessTokenSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to post tweet: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Tweet posted successfully!")
}

func readTweetText() string {
	// Check command line args first
	if len(os.Args) > 1 {
		return strings.Join(os.Args[1:], " ")
	}

	// Check if stdin has data (piped input)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}

	return ""
}

func postTweet(text, consumerKey, consumerSecret, accessToken, accessTokenSecret string) error {
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := postTweetV2(text, consumerKey, consumerSecret, accessToken, accessTokenSecret)
		if err == nil {
			return nil
		}
		if strings.Contains(err.Error(), "503") && attempt < maxRetries {
			wait := time.Duration(attempt*2) * time.Second
			fmt.Fprintf(os.Stderr, "Attempt %d: API returned 503, retrying in %v...\n", attempt, wait)
			time.Sleep(wait)
			continue
		}
		return err
	}
	return fmt.Errorf("all %d attempts failed", maxRetries)
}

func postTweetV2(text, consumerKey, consumerSecret, accessToken, accessTokenSecret string) error {
	body := map[string]string{"text": text}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	req, err := http.NewRequest("POST", tweetURLv2, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	authHeader := buildOAuthHeader("POST", tweetURLv2, nil, consumerKey, consumerSecret, accessToken, accessTokenSecret)
	req.Header.Set("Authorization", authHeader)

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
		Data struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && result.Data.ID != "" {
		fmt.Printf("Tweet ID: %s\n", result.Data.ID)
		fmt.Printf("URL: https://x.com/i/status/%s\n", result.Data.ID)
	}

	return nil
}

func postTweetV1(text, consumerKey, consumerSecret, accessToken, accessTokenSecret string) error {
	// v1.1 uses form-encoded body; status param must be included in OAuth signature
	bodyParams := map[string]string{"status": text}

	form := url.Values{}
	form.Set("status", text)

	req, err := http.NewRequest("POST", tweetURLv1, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	authHeader := buildOAuthHeader("POST", tweetURLv1, bodyParams, consumerKey, consumerSecret, accessToken, accessTokenSecret)
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API v1.1 returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		IDStr string `json:"id_str"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && result.IDStr != "" {
		fmt.Printf("Tweet ID: %s\n", result.IDStr)
		fmt.Printf("URL: https://x.com/i/status/%s\n", result.IDStr)
	}

	return nil
}

// buildOAuthHeader constructs the OAuth 1.0a Authorization header
func buildOAuthHeader(method, apiURL string, bodyParams map[string]string, consumerKey, consumerSecret, accessToken, accessTokenSecret string) string {
	nonce := generateNonce()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	params := map[string]string{
		"oauth_consumer_key":     consumerKey,
		"oauth_nonce":            nonce,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        timestamp,
		"oauth_token":            accessToken,
		"oauth_version":          "1.0",
	}

	// Build signature base string — include body params for form-encoded requests (v1.1)
	allParams := make(map[string]string)
	for k, v := range params {
		allParams[k] = v
	}
	for k, v := range bodyParams {
		allParams[k] = v
	}

	paramPairs := make([]string, 0, len(allParams))
	for k, v := range allParams {
		paramPairs = append(paramPairs, percentEncode(k)+"="+percentEncode(v))
	}
	sort.Strings(paramPairs)

	paramString := strings.Join(paramPairs, "&")
	baseString := method + "&" + percentEncode(apiURL) + "&" + percentEncode(paramString)

	// Sign
	signingKey := percentEncode(consumerSecret) + "&" + percentEncode(accessTokenSecret)
	signature := signHMACSHA1(baseString, signingKey)

	params["oauth_signature"] = signature

	// Build header
	headerPairs := make([]string, 0, len(params))
	for k, v := range params {
		headerPairs = append(headerPairs, percentEncode(k)+`="`+percentEncode(v)+`"`)
	}
	sort.Strings(headerPairs)

	return "OAuth " + strings.Join(headerPairs, ", ")
}

func generateNonce() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func safePrefix(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}

func signHMACSHA1(base, key string) string {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(base))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func percentEncode(s string) string {
	encoded := url.QueryEscape(s)
	// OAuth requires ~ not be encoded, and spaces as %20 not +
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}
