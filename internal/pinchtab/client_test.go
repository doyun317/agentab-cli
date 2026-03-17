package pinchtab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientShutdownTreatsServerExitAsSuccess(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("POST /shutdown", func(w http.ResponseWriter, r *http.Request) {
		go server.Close()
	})

	client := NewClient(server.URL, "", time.Second)
	if err := client.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}
