package main

import (
	"encoding/json"
	"fmt"
	"log"
)

/*
🏷️ JSON STRUCT TAGS - MASTER CLASS

This file shows you EXACTLY how struct tags work for JSON mapping.

Run this file to understand:
1. Why tags are needed
2. How matching works
3. When to use omitempty
4. Common mistakes
5. Real-world patterns
*/

func main() {
	fmt.Println("🏷️ JSON Struct Tags Practice\n")

	example1_WhyTagsAreNeeded()
	example2_HowMatchingWorks()
	example3_OmitEmpty()
	example4_CommonMistakes()
	example5_RealWorldPatterns()
}

// ============================================================
// EXAMPLE 1: Why Tags Are Needed
// ============================================================

func example1_WhyTagsAreNeeded() {
	fmt.Println("============================================================")
	fmt.Println("EXAMPLE 1: Why Struct Tags Are Needed")
	fmt.Println("============================================================")

	// Without tags
	type UserWithoutTags struct {
		UserID int
		Name   string
	}

	// With tags
	type UserWithTags struct {
		UserID int    `json:"user_id"`
		Name   string `json:"name"`
	}

	// JSON from API (typical snake_case)
	jsonData := `{"user_id": 101, "name": "Alice"}`

	// Try WITHOUT tags
	fmt.Println("\n📌 WITHOUT TAGS:")
	fmt.Println("JSON:", jsonData)

	var user1 UserWithoutTags
	json.Unmarshal([]byte(jsonData), &user1)

	fmt.Printf("Result: UserID=%d, Name=%s\n", user1.UserID, user1.Name)
	fmt.Println("❌ UserID is 0! Go looked for 'UserID' but JSON has 'user_id'")

	// Try WITH tags
	fmt.Println("\n📌 WITH TAGS:")
	fmt.Println("JSON:", jsonData)

	var user2 UserWithTags
	json.Unmarshal([]byte(jsonData), &user2)

	fmt.Printf("Result: UserID=%d, Name=%s\n", user2.UserID, user2.Name)
	fmt.Println("✅ Works! Tag tells Go: 'user_id' in JSON → UserID field in struct")

	fmt.Println("\n💡 LESSON: Tags map JSON keys to Go field names")
	fmt.Println()
}

// ============================================================
// EXAMPLE 2: How Matching Works
// ============================================================

func example2_HowMatchingWorks() {
	fmt.Println("============================================================")
	fmt.Println("EXAMPLE 2: How JSON Matching Works")
	fmt.Println("============================================================")

	type Product struct {
		ID       int     `json:"product_id"`  // Maps "product_id" → ID
		Name     string  `json:"productName"` // Maps "productName" → Name (camelCase)
		Price    float64 `json:"price"`       // Maps "price" → Price (exact match)
		Quantity int     // No tag - Go will look for "Quantity" (exact field name)
	}

	// Test different JSON formats
	testCases := []string{
		`{"product_id": 1, "productName": "Laptop", "price": 999.99, "Quantity": 5}`,
		`{"product_id": 2, "productName": "Mouse", "price": 29.99, "quantity": 10}`,
	}

	fmt.Println("\n📌 Struct Definition:")
	fmt.Println("   ID       int    `json:\"product_id\"`")
	fmt.Println("   Name     string `json:\"productName\"`")
	fmt.Println("   Price    float64 `json:\"price\"`")
	fmt.Println("   Quantity int     // NO TAG")

	for i, jsonData := range testCases {
		fmt.Printf("\n🧪 Test Case %d:\n", i+1)
		fmt.Println("JSON:", jsonData)

		var product Product
		json.Unmarshal([]byte(jsonData), &product)

		fmt.Printf("Result: ID=%d, Name=%s, Price=%.2f, Quantity=%d\n",
			product.ID, product.Name, product.Price, product.Quantity)

		if product.Quantity == 0 && i == 1 {
			fmt.Println("⚠️  Quantity is 0! JSON has 'quantity' but Go looks for 'Quantity' (case-sensitive)")
		}
	}

	fmt.Println("\n💡 LESSON: Tag value must EXACTLY match JSON key (case-sensitive)")
	fmt.Println()
}

// ============================================================
// EXAMPLE 3: omitempty - When to Use It
// ============================================================

func example3_OmitEmpty() {
	fmt.Println("============================================================")
	fmt.Println("EXAMPLE 3: The 'omitempty' Tag")
	fmt.Println("============================================================")

	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		ID      int64  `json:"id,omitempty"`    // Omit if 0
		Error   string `json:"error,omitempty"` // Omit if empty string
	}

	// Success response (ID exists)
	successResponse := Response{
		Success: true,
		Message: "Entry created",
		ID:      123,
	}

	// Error response (ID is 0, should be omitted)
	errorResponse := Response{
		Success: false,
		Message: "Validation failed",
		Error:   "user_id must be positive",
	}

	fmt.Println("\n📌 SUCCESS RESPONSE (ID=123):")
	jsonBytes, _ := json.MarshalIndent(successResponse, "", "  ")
	fmt.Println(string(jsonBytes))
	fmt.Println("✅ 'id' is included because it's 123")

	fmt.Println("\n📌 ERROR RESPONSE (ID=0):")
	jsonBytes, _ = json.MarshalIndent(errorResponse, "", "  ")
	fmt.Println(string(jsonBytes))
	fmt.Println("✅ 'id' is omitted because it's 0 (and has omitempty)")
	fmt.Println("✅ 'error' is included because it has a value")

	fmt.Println("\n💡 LESSON: Use omitempty for optional fields")
	fmt.Println("   - Success responses: include ID")
	fmt.Println("   - Error responses: omit ID")
	fmt.Println()
}

// ============================================================
// EXAMPLE 4: Common Mistakes
// ============================================================

func example4_CommonMistakes() {
	fmt.Println("============================================================")
	fmt.Println("EXAMPLE 4: Common Mistakes")
	fmt.Println("============================================================")

	// Mistake 1: Forgetting tags
	type BadUser1 struct {
		UserID int    // ❌ No tag - Go looks for "UserID"
		Name   string // ❌ No tag - Go looks for "Name"
	}

	// Mistake 2: Wrong case in tag
	type BadUser2 struct {
		UserID int    `json:"UserID"` // ❌ Wrong - JSON has "user_id"
		Name   string `json:"name"`
	}

	// Mistake 3: Tag doesn't match API
	type BadUser3 struct {
		UserID int    `json:"id"` // ❌ Wrong - JSON has "user_id"
		Name   string `json:"name"`
	}

	// Correct version
	type GoodUser struct {
		UserID int    `json:"user_id"` // ✅ Matches JSON exactly
		Name   string `json:"name"`
	}

	jsonData := `{"user_id": 101, "name": "Alice"}`

	fmt.Println("\n📌 JSON from API:")
	fmt.Println(jsonData)

	// Test all versions
	fmt.Println("\n❌ Mistake 1: No tags")
	var bad1 BadUser1
	json.Unmarshal([]byte(jsonData), &bad1)
	fmt.Printf("   Result: UserID=%d (WRONG - expected 101)\n", bad1.UserID)

	fmt.Println("\n❌ Mistake 2: Wrong case")
	var bad2 BadUser2
	json.Unmarshal([]byte(jsonData), &bad2)
	fmt.Printf("   Result: UserID=%d (WRONG - expected 101)\n", bad2.UserID)

	fmt.Println("\n❌ Mistake 3: Tag doesn't match JSON key")
	var bad3 BadUser3
	json.Unmarshal([]byte(jsonData), &bad3)
	fmt.Printf("   Result: UserID=%d (WRONG - expected 101)\n", bad3.UserID)

	fmt.Println("\n✅ Correct: Tag matches JSON exactly")
	var good GoodUser
	json.Unmarshal([]byte(jsonData), &good)
	fmt.Printf("   Result: UserID=%d (CORRECT!)\n", good.UserID)

	fmt.Println("\n💡 LESSON: Tag value must EXACTLY match the JSON key")
	fmt.Println()
}

// ============================================================
// EXAMPLE 5: Real-World Patterns
// ============================================================

func example5_RealWorldPatterns() {
	fmt.Println("============================================================")
	fmt.Println("EXAMPLE 5: Real-World API Patterns")
	fmt.Println("============================================================")

	// Pattern 1: REST API Request (snake_case)
	type CreateEntryRequest struct {
		UserID   int    `json:"user_id"`
		Text     string `json:"text"`
		Mood     int    `json:"mood"`
		Category string `json:"category"`
	}

	// Pattern 2: External API (camelCase)
	type GitHubUser struct {
		Login     string `json:"login"`
		ID        int    `json:"id"`
		AvatarURL string `json:"avatar_url"` // Maps "avatar_url" → AvatarURL
		Name      string `json:"name"`
	}

	// Pattern 3: Response with optional fields
	type APIResponse struct {
		Success bool        `json:"success"`
		Message string      `json:"message"`
		Data    interface{} `json:"data,omitempty"`
		Error   string      `json:"error,omitempty"`
	}

	fmt.Println("\n📌 Pattern 1: REST API (snake_case)")
	reqJSON := `{"user_id": 101, "text": "Feeling great!", "mood": 8, "category": "work"}`
	var req CreateEntryRequest
	json.Unmarshal([]byte(reqJSON), &req)
	fmt.Println("   JSON:", reqJSON)
	fmt.Printf("   Parsed: UserID=%d, Text=%s, Mood=%d\n", req.UserID, req.Text, req.Mood)

	fmt.Println("\n📌 Pattern 2: External API (camelCase)")
	githubJSON := `{"login": "octocat", "id": 583231, "avatar_url": "https://...", "name": "The Octocat"}`
	var ghUser GitHubUser
	json.Unmarshal([]byte(githubJSON), &ghUser)
	fmt.Println("   JSON:", githubJSON)
	fmt.Printf("   Parsed: Login=%s, AvatarURL=%s\n", ghUser.Login, ghUser.AvatarURL)

	fmt.Println("\n📌 Pattern 3: API Response (conditional fields)")
	successResp := APIResponse{Success: true, Message: "OK", Data: map[string]int{"count": 5}}
	successJSON, _ := json.MarshalIndent(successResp, "   ", "  ")
	fmt.Println("   Success response:")
	fmt.Println("  ", string(successJSON))

	errorResp := APIResponse{Success: false, Message: "Failed", Error: "Invalid input"}
	errorJSON, _ := json.MarshalIndent(errorResp, "   ", "  ")
	fmt.Println("\n   Error response:")
	fmt.Println("  ", string(errorJSON))

	fmt.Println("\n💡 LESSON: Different APIs use different conventions")
	fmt.Println("   - Your API: Use consistent naming (snake_case or camelCase)")
	fmt.Println("   - External APIs: Match their exact convention in tags")
	fmt.Println()
}

/*
🎓 QUICK REFERENCE CARD:

1. BASIC TAG:
   type User struct {
       ID int `json:"user_id"`
   }
   Maps: "user_id" in JSON → ID field in struct

2. OMITEMPTY:
   type Response struct {
       ID int64 `json:"id,omitempty"`
   }
   Omits field from JSON if it's zero value (0, "", false, nil)

3. IGNORE FIELD:
   type User struct {
       Password string `json:"-"`
   }
   Never include this field in JSON

4. MATCHING RULES:
   - Tag value must EXACTLY match JSON key (case-sensitive)
   - Without tag, Go uses exact field name
   - "user_id" ≠ "UserID" ≠ "userId"

5. COMMON CONVENTIONS:
   - REST APIs: snake_case (user_id, created_at)
   - JavaScript APIs: camelCase (userId, createdAt)
   - Go structs: PascalCase (UserID, CreatedAt)

6. INTERVIEW ANSWER:
   "Struct tags provide metadata for JSON marshaling. The json tag
    maps between Go field names and JSON keys, handling different
    naming conventions. omitempty excludes zero values from output."

🔥 PRACTICE:
   1. Run: go run learning-demos/02-json-struct-tags-practice.go
   2. Modify examples and re-run
   3. Try your own struct with different APIs
*/

func init() {
	log.SetFlags(0) // Remove timestamp from logs for cleaner output
}
