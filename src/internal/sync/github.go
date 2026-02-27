package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultBaseURL = "https://api.github.com"

// GistClient communicates with the GitHub Gist API.
type GistClient struct {
	token   string
	baseURL string
	http    *http.Client
}

// NewGistClient creates a new GitHub Gist API client.
func NewGistClient(token string) *GistClient {
	return &GistClient{
		token:   token,
		baseURL: defaultBaseURL,
		http:    &http.Client{},
	}
}

// Gist represents a GitHub Gist.
type Gist struct {
	ID          string              `json:"id"`
	Description string              `json:"description"`
	Files       map[string]GistFile `json:"files"`
	UpdatedAt   string              `json:"updated_at"`
	HTMLURL     string              `json:"html_url"`
}

// GistFile represents a file within a Gist.
type GistFile struct {
	Filename string `json:"filename,omitempty"`
	Content  string `json:"content"`
	Size     int    `json:"size,omitempty"`
}

// GistUpdate is the payload for updating a Gist.
type GistUpdate struct {
	Description string               `json:"description,omitempty"`
	Files       map[string]*GistFile `json:"files"`
}

func (c *GistClient) do(method, path string, body interface{}) (*http.Response, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.http.Do(req)
}

// ValidateToken checks the token by calling GET /user and returns the username.
func (c *GistClient) ValidateToken() (string, error) {
	resp, err := c.do("GET", "/user", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("invalid token (status %d)", resp.StatusCode)
	}

	var user struct {
		Login string `json:"login"`
	}
	json.NewDecoder(resp.Body).Decode(&user)
	return user.Login, nil
}

// GetGist retrieves a Gist by ID.
func (c *GistClient) GetGist(id string) (*Gist, error) {
	resp, err := c.do("GET", "/gists/"+id, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api error %d: %s", resp.StatusCode, string(body))
	}

	var gist Gist
	json.NewDecoder(resp.Body).Decode(&gist)
	return &gist, nil
}

// CreateGist creates a new Gist.
func (c *GistClient) CreateGist(description string, files map[string]GistFile, public bool) (*Gist, error) {
	payload := map[string]interface{}{
		"description": description,
		"public":      public,
		"files":       files,
	}

	resp, err := c.do("POST", "/gists", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api error %d: %s", resp.StatusCode, string(body))
	}

	var gist Gist
	json.NewDecoder(resp.Body).Decode(&gist)
	return &gist, nil
}

// UpdateGist updates an existing Gist.
func (c *GistClient) UpdateGist(id string, update GistUpdate) (*Gist, error) {
	resp, err := c.do("PATCH", "/gists/"+id, update)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api error %d: %s", resp.StatusCode, string(body))
	}

	var gist Gist
	json.NewDecoder(resp.Body).Decode(&gist)
	return &gist, nil
}
