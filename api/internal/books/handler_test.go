package books

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSearchBooks(t *testing.T) {
	// Mock Open Library API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify query parameters
		// The handler sends 'title' as the query parameter for 'q'
		q := r.URL.Query().Get("title")
		if q == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Return mock response
		resp := olResponse{
			NumFound: 1,
			Docs: []olDoc{
				{
					Key:              "/works/OL123W",
					Title:            "Test Book",
					AuthorName:       []string{"Test Author"},
					FirstPublishYear: ptr(2023),
					EditionCount:     1,
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	// Override external URL
	originalURL := olSearchURL
	olSearchURL = ts.URL
	defer func() { olSearchURL = originalURL }()

	// Create Handler (nil pool and search client for unit test)
	h := NewHandler(nil, nil)

	// Setup Gin
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/books/search?q=test", nil)

	// Execute
	h.SearchBooks(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, float64(1), result["total"])
	docs := result["results"].([]interface{})
	assert.Len(t, docs, 1)

	doc := docs[0].(map[string]interface{})
	assert.Equal(t, "/works/OL123W", doc["key"])
	assert.Equal(t, "Test Book", doc["title"])
}

func TestGetBook(t *testing.T) {
	// Mock Open Library API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/works/OL123W.json":
			json.NewEncoder(w).Encode(olWork{
				Title: "Test Book",
				Description: json.RawMessage(`"Test Description"`),
				Authors: []olAuthorRef{{Author: struct{Key string `json:"key"`}{Key: "/authors/OL1A"}}},
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

	// Override external URL
	originalURL := olBaseURL
	olBaseURL = ts.URL
	defer func() { olBaseURL = originalURL }()

	h := NewHandler(nil, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "workId", Value: "OL123W"}}
	c.Request, _ = http.NewRequest("GET", "/books/OL123W", nil)

	h.GetBook(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result BookDetail
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, "Test Book", result.Title)
	assert.Equal(t, "Test Description", *result.Description)
	assert.Contains(t, result.Authors, "Test Author")
}

func ptr(i int) *int {
	return &i
}
