package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	baseURL := os.Getenv("API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Test de création de conversation
	fmt.Println("🔄 Testing conversation creation...")
	testCreateConversation(baseURL)

	// Test d'envoi de message
	fmt.Println("🔄 Testing message sending...")
	testSendMessage(baseURL)

	// Test d'envoi de message payant
	fmt.Println("🔄 Testing paid message sending...")
	testSendPaidMessage(baseURL)

	// Test de récupération des conversations
	fmt.Println("🔄 Testing conversation listing...")
	testGetConversations(baseURL)
}

func testCreateConversation(baseURL string) {
	payload := map[string]interface{}{
		"other_user_id": "user-2",
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := makeAuthenticatedRequest("POST", baseURL+"/api/v1/conversations", jsonData)
	if err != nil {
		fmt.Printf("❌ Error creating conversation: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✅ Create conversation response: %d\n", resp.StatusCode)
	} else {
		fmt.Printf("❌ Create conversation failed: %d\n", resp.StatusCode)
	}
}

func testSendMessage(baseURL string) {
	content := "Hello from API test!"
	payload := map[string]interface{}{
		"conversation_id": 1,
		"content":         &content,
		"message_type":    "text",
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := makeAuthenticatedRequest("POST", baseURL+"/api/v1/messages", jsonData)
	if err != nil {
		fmt.Printf("❌ Error sending message: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("✅ Send message response: %d\n", resp.StatusCode)
	} else {
		fmt.Printf("❌ Send message failed: %d\n", resp.StatusCode)
	}
}

func testSendPaidMessage(baseURL string) {
	content := "Premium content!"
	payload := map[string]interface{}{
		"conversation_id": 1,
		"content":         &content,
		"price":           9.99,
		"message_type":    "paid_text",
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := makeAuthenticatedRequest("POST", baseURL+"/api/v1/messages/paid", jsonData)
	if err != nil {
		fmt.Printf("❌ Error sending paid message: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("✅ Send paid message response: %d\n", resp.StatusCode)
	} else {
		fmt.Printf("❌ Send paid message failed: %d\n", resp.StatusCode)
	}
}

func testGetConversations(baseURL string) {
	resp, err := makeAuthenticatedRequest("GET", baseURL+"/api/v1/conversations", nil)
	if err != nil {
		fmt.Printf("❌ Error getting conversations: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✅ Get conversations response: %d\n", resp.StatusCode)
	} else {
		fmt.Printf("❌ Get conversations failed: %d\n", resp.StatusCode)
	}
}

func makeAuthenticatedRequest(method, url string, data []byte) (*http.Response, error) {
	var req *http.Request
	var err error

	if data != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(data))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	// Ajouter les headers
	req.Header.Set("Content-Type", "application/json")
	// TODO: Ajouter un token JWT valide pour les tests
	// req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}
