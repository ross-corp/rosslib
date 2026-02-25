package genreratings

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tristansaldanha/rosslib/api/internal/middleware"
)

// allowedGenres is the canonical set of genre dimensions users can rate.
// Kept in sync with the predefined genre list in books/handler.go.
var allowedGenres = map[string]bool{
	"Fiction":         true,
	"Non-fiction":     true,
	"Fantasy":         true,
	"Science fiction": true,
	"Mystery":         true,
	"Romance":         true,
	"Horror":          true,
	"Thriller":        true,
	"Biography":       true,
	"History":         true,
	"Poetry":          true,
	"Children":        true,
}

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

type genreRating struct {
	Genre     string `json:"genre"`
	Rating    int    `json:"rating"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type aggregateGenreRating struct {
	Genre      string  `json:"genre"`
	Average    float64 `json:"average"`
	RaterCount int     `json:"rater_count"`
}

// GetBookGenreRatings returns aggregate genre ratings for a book.
// GET /books/:workId/genre-ratings
func (h *Handler) GetBookGenreRatings(c *gin.Context) {
	workId := c.Param("workId")

	var bookID string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id FROM books WHERE open_library_id = $1`, workId,
	).Scan(&bookID)
	if err != nil {
		c.JSON(http.StatusOK, []aggregateGenreRating{})
		return
	}

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT genre, AVG(rating)::float8, COUNT(*)
		 FROM genre_ratings
		 WHERE book_id = $1
		 GROUP BY genre
		 ORDER BY COUNT(*) DESC, genre`, bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	results := []aggregateGenreRating{}
	for rows.Next() {
		var g aggregateGenreRating
		if err := rows.Scan(&g.Genre, &g.Average, &g.RaterCount); err != nil {
			continue
		}
		results = append(results, g)
	}

	c.JSON(http.StatusOK, results)
}

// GetMyGenreRatings returns the current user's genre ratings for a book.
// GET /me/books/:olId/genre-ratings
func (h *Handler) GetMyGenreRatings(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	olID := c.Param("olId")

	var bookID string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id FROM books WHERE open_library_id = $1`, olID,
	).Scan(&bookID)
	if err != nil {
		c.JSON(http.StatusOK, []genreRating{})
		return
	}

	rows, err := h.pool.Query(c.Request.Context(),
		`SELECT genre, rating, updated_at
		 FROM genre_ratings
		 WHERE user_id = $1 AND book_id = $2
		 ORDER BY genre`, userID, bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	defer rows.Close()

	results := []genreRating{}
	for rows.Next() {
		var g genreRating
		if err := rows.Scan(&g.Genre, &g.Rating, &g.UpdatedAt); err != nil {
			continue
		}
		results = append(results, g)
	}

	c.JSON(http.StatusOK, results)
}

// SetGenreRatings sets or updates the current user's genre ratings for a book.
// Accepts an array of {genre, rating} objects. Ratings with value 0 or missing
// genres are deleted (removing a genre rating). Genres must be from the allowed set.
// PUT /me/books/:olId/genre-ratings
func (h *Handler) SetGenreRatings(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	olID := c.Param("olId")

	var input []struct {
		Genre  string `json:"genre"`
		Rating *int   `json:"rating"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expected array of {genre, rating}"})
		return
	}

	// Validate genres and ratings
	for _, item := range input {
		genre := strings.TrimSpace(item.Genre)
		if !allowedGenres[genre] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre: " + genre})
			return
		}
		if item.Rating != nil && (*item.Rating < 0 || *item.Rating > 10) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be 0-10"})
			return
		}
	}

	// Look up book
	var bookID string
	err := h.pool.QueryRow(c.Request.Context(),
		`SELECT id FROM books WHERE open_library_id = $1`, olID,
	).Scan(&bookID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	// Upsert each rating (or delete if rating is null/0)
	for _, item := range input {
		genre := strings.TrimSpace(item.Genre)
		if item.Rating == nil || *item.Rating == 0 {
			// Delete this genre rating
			_, _ = h.pool.Exec(c.Request.Context(),
				`DELETE FROM genre_ratings WHERE user_id = $1 AND book_id = $2 AND genre = $3`,
				userID, bookID, genre)
		} else {
			_, err := h.pool.Exec(c.Request.Context(),
				`INSERT INTO genre_ratings (user_id, book_id, genre, rating)
				 VALUES ($1, $2, $3, $4)
				 ON CONFLICT (user_id, book_id, genre) DO UPDATE
				 SET rating = $4, updated_at = NOW()`,
				userID, bookID, genre, *item.Rating)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
