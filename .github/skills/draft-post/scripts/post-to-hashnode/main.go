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

const hashnodeAPI = "https://gql.hashnode.com"

func main() {
	_ = godotenv.Load()

	token := os.Getenv("HASHNODE_TOKEN")
	pubID := os.Getenv("HASHNODE_PUBLICATION_ID")

	if token == "" {
		fmt.Fprintln(os.Stderr, "Error: HASHNODE_TOKEN not set. Add it to .env")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  go run main.go me                              Get your user info + publication IDs")
		fmt.Fprintln(os.Stderr, "  go run main.go draft <markdown-file>            Create draft")
		fmt.Fprintln(os.Stderr, "  go run main.go publish <markdown-file>          Publish article")
		fmt.Fprintln(os.Stderr, "  go run main.go list                             List your posts")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Markdown file should have YAML frontmatter:")
		fmt.Fprintln(os.Stderr, "  ---")
		fmt.Fprintln(os.Stderr, "  title: My Article Title")
		fmt.Fprintln(os.Stderr, "  tags: go, buildinpublic")
		fmt.Fprintln(os.Stderr, "  ---")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "me":
		err := getMe(token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
			os.Exit(1)
		}
	case "draft":
		if pubID == "" {
			fmt.Fprintln(os.Stderr, "Error: HASHNODE_PUBLICATION_ID not set. Run 'go run main.go me' to find it, then add to .env")
			os.Exit(1)
		}
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: provide a markdown file path")
			os.Exit(1)
		}
		err := createDraft(token, pubID, os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
			os.Exit(1)
		}
	case "publish":
		if pubID == "" {
			fmt.Fprintln(os.Stderr, "Error: HASHNODE_PUBLICATION_ID not set. Run 'go run main.go me' to find it, then add to .env")
			os.Exit(1)
		}
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: provide a markdown file path")
			os.Exit(1)
		}
		err := publishPost(token, pubID, os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
			os.Exit(1)
		}
	case "list":
		if pubID == "" {
			fmt.Fprintln(os.Stderr, "Error: HASHNODE_PUBLICATION_ID not set. Run 'go run main.go me' to find it, then add to .env")
			os.Exit(1)
		}
		err := listPosts(token, pubID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

type gqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

func gqlCall(token string, req gqlRequest) (map[string]interface{}, error) {
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	httpReq, err := http.NewRequest("POST", hashnodeAPI, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if errors, ok := result["errors"]; ok {
		errJSON, _ := json.MarshalIndent(errors, "", "  ")
		return nil, fmt.Errorf("GraphQL errors: %s", string(errJSON))
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response: %s", string(respBody))
	}

	return data, nil
}

func getMe(token string) error {
	query := `query {
		me {
			id
			username
			name
			publications(first: 5) {
				edges {
					node {
						id
						title
						url
					}
				}
			}
		}
	}`

	data, err := gqlCall(token, gqlRequest{Query: query})
	if err != nil {
		return err
	}

	me, _ := data["me"].(map[string]interface{})
	fmt.Printf("User: %s (@%s)\n", me["name"], me["username"])
	fmt.Printf("ID: %s\n\n", me["id"])

	pubs, _ := me["publications"].(map[string]interface{})
	edges, _ := pubs["edges"].([]interface{})

	fmt.Println("Your Publications:")
	fmt.Printf("%-30s %s\n", "ID (add to .env)", "Title")
	fmt.Println(strings.Repeat("-", 70))
	for _, edge := range edges {
		e, _ := edge.(map[string]interface{})
		node, _ := e["node"].(map[string]interface{})
		fmt.Printf("%-30s %s\n", node["id"], node["title"])
		fmt.Printf("%-30s %s\n", "", node["url"])
	}

	return nil
}

func publishPost(token, pubID, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	title, tags, _, body := parseFrontmatter(string(data))
	if title == "" {
		return fmt.Errorf("markdown file must have 'title' in frontmatter")
	}

	tagInput := buildTagInput(tags)

	query := `mutation PublishPost($input: PublishPostInput!) {
		publishPost(input: $input) {
			post {
				id
				title
				url
			}
		}
	}`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"title":           title,
			"contentMarkdown": body,
			"publicationId":   pubID,
			"tags":            tagInput,
		},
	}

	result, err := gqlCall(token, gqlRequest{Query: query, Variables: variables})
	if err != nil {
		return err
	}

	pub, _ := result["publishPost"].(map[string]interface{})
	post, _ := pub["post"].(map[string]interface{})
	fmt.Println("Article PUBLISHED!")
	fmt.Printf("ID: %s\n", post["id"])
	fmt.Printf("URL: %s\n", post["url"])

	return nil
}

func createDraft(token, pubID, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	title, tags, _, body := parseFrontmatter(string(data))
	if title == "" {
		return fmt.Errorf("markdown file must have 'title' in frontmatter")
	}

	tagInput := buildTagInput(tags)

	query := `mutation CreateDraft($input: CreateDraftInput!) {
		createDraft(input: $input) {
			draft {
				id
				title
			}
		}
	}`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"title":           title,
			"contentMarkdown": body,
			"publicationId":   pubID,
			"tags":            tagInput,
		},
	}

	result, err := gqlCall(token, gqlRequest{Query: query, Variables: variables})
	if err != nil {
		return err
	}

	draft, _ := result["createDraft"].(map[string]interface{})
	d, _ := draft["draft"].(map[string]interface{})
	fmt.Println("Draft CREATED!")
	fmt.Printf("ID: %s\n", d["id"])
	fmt.Printf("Title: %s\n", d["title"])
	fmt.Println("Review at: https://hashnode.com/drafts")

	return nil
}

func listPosts(token, pubID string) error {
	query := `query ListPosts($pubId: ObjectId!, $first: Int!) {
		publication(id: $pubId) {
			posts(first: $first) {
				edges {
					node {
						id
						title
						url
						publishedAt
					}
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"pubId": pubID,
		"first": 10,
	}

	data, err := gqlCall(token, gqlRequest{Query: query, Variables: variables})
	if err != nil {
		return err
	}

	pub, _ := data["publication"].(map[string]interface{})
	posts, _ := pub["posts"].(map[string]interface{})
	edges, _ := posts["edges"].([]interface{})

	if len(edges) == 0 {
		fmt.Println("No posts yet. Start writing!")
		return nil
	}

	fmt.Printf("%-8s %s\n", "ID", "Title")
	fmt.Println(strings.Repeat("-", 60))
	for _, edge := range edges {
		e, _ := edge.(map[string]interface{})
		node, _ := e["node"].(map[string]interface{})
		fmt.Printf("%-8s %s\n", node["id"], node["title"])
	}

	return nil
}

func buildTagInput(tags []string) []map[string]string {
	var tagInput []map[string]string
	for _, t := range tags {
		tagInput = append(tagInput, map[string]string{
			"slug": strings.ToLower(strings.ReplaceAll(t, " ", "-")),
			"name": t,
		})
	}
	return tagInput
}

func parseFrontmatter(content string) (title string, tags []string, series string, body string) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return "", nil, "", content
	}

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
