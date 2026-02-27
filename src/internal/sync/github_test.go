package sync

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateToken_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing or wrong Authorization header")
		}
		json.NewEncoder(w).Encode(map[string]string{"login": "testuser"})
	}))
	defer srv.Close()

	client := NewGistClient("test-token")
	client.baseURL = srv.URL

	username, err := client.ValidateToken()
	if err != nil {
		t.Fatalf("ValidateToken() error: %v", err)
	}
	if username != "testuser" {
		t.Errorf("username = %q, want %q", username, "testuser")
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer srv.Close()

	client := NewGistClient("bad-token")
	client.baseURL = srv.URL

	_, err := client.ValidateToken()
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestGetGist(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gists/gist123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("unexpected method: %s", r.Method)
		}
		json.NewEncoder(w).Encode(Gist{
			ID:          "gist123",
			Description: "snipt-sync",
			Files: map[string]GistFile{
				"test.md": {Filename: "test.md", Content: "hello", Size: 5},
			},
		})
	}))
	defer srv.Close()

	client := NewGistClient("token")
	client.baseURL = srv.URL

	gist, err := client.GetGist("gist123")
	if err != nil {
		t.Fatalf("GetGist() error: %v", err)
	}
	if gist.ID != "gist123" {
		t.Errorf("ID = %q, want %q", gist.ID, "gist123")
	}
	if len(gist.Files) != 1 {
		t.Errorf("Files count = %d, want 1", len(gist.Files))
	}
	if gist.Files["test.md"].Content != "hello" {
		t.Errorf("File content = %q, want %q", gist.Files["test.md"].Content, "hello")
	}
}

func TestCreateGist(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gists" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)

		if payload["public"] != false {
			t.Error("expected public=false")
		}

		w.WriteHeader(201)
		json.NewEncoder(w).Encode(Gist{
			ID:          "new-gist",
			Description: payload["description"].(string),
		})
	}))
	defer srv.Close()

	client := NewGistClient("token")
	client.baseURL = srv.URL

	gist, err := client.CreateGist("snipt-sync", map[string]GistFile{
		".snipt-meta.json": {Content: "{}"},
	}, false)
	if err != nil {
		t.Fatalf("CreateGist() error: %v", err)
	}
	if gist.ID != "new-gist" {
		t.Errorf("ID = %q, want %q", gist.ID, "new-gist")
	}
}

func TestUpdateGist(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gists/gist123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("unexpected method: %s", r.Method)
		}

		json.NewEncoder(w).Encode(Gist{
			ID: "gist123",
			Files: map[string]GistFile{
				"updated.md": {Filename: "updated.md", Content: "new content"},
			},
		})
	}))
	defer srv.Close()

	client := NewGistClient("token")
	client.baseURL = srv.URL

	gist, err := client.UpdateGist("gist123", GistUpdate{
		Files: map[string]*GistFile{
			"updated.md": {Content: "new content"},
		},
	})
	if err != nil {
		t.Fatalf("UpdateGist() error: %v", err)
	}
	if gist.ID != "gist123" {
		t.Errorf("ID = %q, want %q", gist.ID, "gist123")
	}
}

func TestGetGist_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer srv.Close()

	client := NewGistClient("token")
	client.baseURL = srv.URL

	_, err := client.GetGist("nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}
