package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

const baseURL = "http://localhost:8080"

func TestBuyMerch(t *testing.T) {
	//Auth, token
	authRequest := AuthRequest{
		Username: "testuser",
		Password: "testpassword",
	}
	authResponse, err := authenticate(authRequest)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	//Buy merch
	item := "t-shirt"
	req, err := http.NewRequest("GET", baseURL+"/api/buy/"+item, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+*authResponse.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %v", resp.Status)
	}
}
func TestSendCoins(t *testing.T) {
	//Auth, token
	authRequest := AuthRequest{
		Username: "testuser",
		Password: "testpassword",
	}
	authResponse, err := authenticate(authRequest)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}
	//Need to create reciever in order to send coins
	authRequestReciever := AuthRequest{
		Username: "reciever",
		Password: "rcvpass",
	}
	_, err = authenticate(authRequestReciever)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	sendCoinRequest := SendCoinRequest{
		ToUser: "reciever",
		Amount: 10,
	}
	body, err := json.Marshal(sendCoinRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/api/sendCoin", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+*authResponse.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %v", resp.Status)
	}
}
func TestGetInfo(t *testing.T) {
	//Auth, token
	authRequest := AuthRequest{
		Username: "testuser",
		Password: "testpassword",
	}
	authResponse, err := authenticate(authRequest)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	req, err := http.NewRequest("GET", baseURL+"/api/info", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+*authResponse.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 OK, got %v", resp.Status)
	}

	var infoResponse InfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&infoResponse); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if infoResponse.Coins == nil {
		t.Error("Expected coins to be present in the response")
	}
	if infoResponse.Inventory == nil {
		t.Error("Expected inventory to be present in the response")
	}
	if infoResponse.CoinHistory == nil {
		t.Error("Expected coin history to be present in the response")
	}
}
func authenticate(authRequest AuthRequest) (*AuthResponse, error) {
	body, err := json.Marshal(authRequest)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseURL+"/api/auth", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed with status: %v", resp.Status)
	}

	var authResponse AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return nil, err
	}

	return &authResponse, nil
}
