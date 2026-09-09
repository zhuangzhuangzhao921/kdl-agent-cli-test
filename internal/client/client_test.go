package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zhuangzhuangzhao921/kdl-agent-cli-test/internal/client"
	"github.com/zhuangzhuangzhao921/kdl-agent-cli-test/internal/config"
)

func TestGetSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("auth header: %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":    true,
			"request_id": "req_test",
			"data":       map[string]any{"balance": 1.23},
		})
	}))
	defer srv.Close()

	c, err := client.New(config.Resolved{GatewayURL: srv.URL, Token: "test-token"})
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Get(context.Background(), "/v1/account/funds", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Success || resp.RequestID != "req_test" {
		t.Fatalf("resp: %+v", resp)
	}
}

func TestGetAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":    false,
			"request_id": "req_denied",
			"error":      map[string]any{"code": "AUTH_REQUIRED", "message": "需要有效的 Agent 凭证"},
		})
	}))
	defer srv.Close()

	c, err := client.New(config.Resolved{GatewayURL: srv.URL, Token: "bad"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Get(context.Background(), "/v1/account/funds", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want APIError, got %T", err)
	}
	if apiErr.Code != "AUTH_REQUIRED" {
		t.Fatalf("code: %s", apiErr.Code)
	}
}
