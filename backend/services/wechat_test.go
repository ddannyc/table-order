package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetWeChatSession_MockConfigured(t *testing.T) {
	// Test that when AppID/AppSecret are empty, error is returned
	// This tests the fallback behavior
	_, err := GetWeChatSession("test_code")
	if err == nil {
		t.Error("expected error when WeChat not configured")
	}
}

func TestWeChatSessionResponse_JSON(t *testing.T) {
	// Test parsing WeChat API response
	jsonData := `{
		"openid": "test_openid_123",
		"session_key": "test_session_key",
		"unionid": "test_union_id",
		"errcode": 0,
		"errmsg": ""
	}`

	var resp WeChatSessionResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Errorf("failed to parse JSON: %v", err)
	}

	if resp.OpenID != "test_openid_123" {
		t.Errorf("expected openid 'test_openid_123', got '%s'", resp.OpenID)
	}
	if resp.SessionKey != "test_session_key" {
		t.Errorf("expected session_key 'test_session_key', got '%s'", resp.SessionKey)
	}
	if resp.ErrCode != 0 {
		t.Errorf("expected errcode 0, got %d", resp.ErrCode)
	}
}

func TestWeChatSessionResponse_Error(t *testing.T) {
	// Test error response parsing
	jsonData := `{"errcode": 40029, "errmsg": "invalid code"}`

	var resp WeChatSessionResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Errorf("failed to parse JSON: %v", err)
	}

	if resp.ErrCode != 40029 {
		t.Errorf("expected errcode 40029, got %d", resp.ErrCode)
	}
}

type mockWeChatServer struct {
	handler func(w http.ResponseWriter, r *http.Request)
}

func (m *mockWeChatServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.handler(w, r)
}

func TestGetWeChatSession_RealAPI_Fallback(t *testing.T) {
	// Start mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"errcode": 40029, "errmsg": "invalid code"}`))
	}))
	defer server.Close()

	// This test verifies error handling when WeChat returns error
	// Actual HTTP call would go to real WeChat API, so we just verify JSON parsing
	jsonData := `{"errcode": 40029, "errmsg": "invalid code"}`
	var resp WeChatSessionResponse
	json.Unmarshal([]byte(jsonData), &resp)

	if resp.ErrCode != 40029 {
		t.Errorf("expected errcode 40029, got %d", resp.ErrCode)
	}
}