package books

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

// rewriteTransport intercepts outgoing requests and redirects them to a test
// server, preserving the original path and query string. This avoids changing
// the package-level const URLs to var.
type rewriteTransport struct {
	target *url.URL
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.target.Scheme
	req.URL.Host = t.target.Host
	return http.DefaultTransport.RoundTrip(req)
}

func olTestClient(ts *httptest.Server) *http.Client {
	u, _ := url.Parse(ts.URL)
	return &http.Client{Transport: &rewriteTransport{target: u}}
}

func TestSearchBooks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("title")
		if q == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp := olResponse{
			NumFound: 1,
			Docs: []olDoc{
				{
					Key:              "/works/OL123W",
					Title:            "Test Book",
					AuthorName:       []string{"Test Author"},
					FirstPublishYear: intPtr(2023),
					EditionCount:     1,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	h := NewHandler(nil, nil, olTestClient(ts))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/books/search?q=test", nil)

	h.SearchBooks(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	total, ok := result["total"].(float64)
	if !ok || total < 1 {
		t.Errorf("expected total >= 1, got %v", result["total"])
	}

	docs, ok := result["results"].([]interface{})
	if !ok || len(docs) == 0 {
		t.Fatalf("expected at least 1 result, got %v", result["results"])
	}

	doc := docs[0].(map[string]interface{})
	if doc["key"] != "/works/OL123W" {
		t.Errorf("expected key /works/OL123W, got %v", doc["key"])
	}
	if doc["title"] != "Test Book" {
		t.Errorf("expected title Test Book, got %v", doc["title"])
	}
}

func TestGetBook(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/works/OL123W.json":
			json.NewEncoder(w).Encode(olWork{
				Title:       "Test Book",
				Description: json.RawMessage(`"A test description"`),
				Authors:     []olAuthorRef{{Author: struct{ Key string `json:"key"` }{Key: "/authors/OL1A"}}},
			})
		case "/works/OL123W/ratings.json":
			json.NewEncoder(w).Encode(olRatings{})
		case "/works/OL123W/editions.json":
			json.NewEncoder(w).Encode(olEditionsResponse{})
		case "/authors/OL1A.json":
			json.NewEncoder(w).Encode(olAuthor{Name: "Test Author"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	h := NewHandler(nil, nil, olTestClient(ts))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "workId", Value: "OL123W"}}
	c.Request, _ = http.NewRequest("GET", "/books/OL123W", nil)

	h.GetBook(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result BookDetail
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result.Title != "Test Book" {
		t.Errorf("expected title 'Test Book', got %q", result.Title)
	}
	if result.Description == nil || *result.Description != "A test description" {
		t.Errorf("expected description 'A test description', got %v", result.Description)
	}
	found := false
	for _, a := range result.Authors {
		if a == "Test Author" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'Test Author' in authors %v", result.Authors)
	}
}

func intPtr(i int) *int {
	return &i
}
