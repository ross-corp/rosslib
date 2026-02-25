package handlers

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// GetBookGenreRatings handles GET /books/{workId}/genre-ratings
func GetBookGenreRatings(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		workID := e.Request.PathValue("workId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": workID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, []any{})
		}

		type ratingRow struct {
			Genre     string  `db:"genre" json:"genre"`
			AvgRating float64 `db:"avg_rating" json:"avg_rating"`
			Count     int     `db:"count" json:"count"`
		}
		var ratings []ratingRow
		err := app.DB().NewQuery(`
			SELECT genre, AVG(rating) as avg_rating, COUNT(*) as count
			FROM genre_ratings WHERE book = {:book}
			GROUP BY genre
			ORDER BY count DESC
		`).Bind(map[string]any{"book": books[0].Id}).All(&ratings)
		if err != nil || ratings == nil {
			ratings = []ratingRow{}
		}

		return e.JSON(http.StatusOK, ratings)
	}
}

// GetMyGenreRatings handles GET /me/books/{olId}/genre-ratings
func GetMyGenreRatings(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		olID := e.Request.PathValue("olId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": olID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusOK, []any{})
		}

		ratings, err := app.FindRecordsByFilter("genre_ratings",
			"user = {:user} && book = {:book}",
			"", 100, 0,
			map[string]any{"user": user.Id, "book": books[0].Id},
		)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var result []map[string]any
		for _, r := range ratings {
			result = append(result, map[string]any{
				"genre":  r.GetString("genre"),
				"rating": r.GetFloat("rating"),
			})
		}
		if result == nil {
			result = []map[string]any{}
		}

		return e.JSON(http.StatusOK, result)
	}
}

// SetGenreRatings handles PUT /me/books/{olId}/genre-ratings
func SetGenreRatings(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}
		olID := e.Request.PathValue("olId")

		books, _ := app.FindRecordsByFilter("books",
			"open_library_id = {:id}", "", 1, 0,
			map[string]any{"id": olID},
		)
		if len(books) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Book not found"})
		}
		book := books[0]

		data := struct {
			Ratings []struct {
				Genre  string  `json:"genre"`
				Rating float64 `json:"rating"`
			} `json:"ratings"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		// Delete existing ratings
		existing, _ := app.FindRecordsByFilter("genre_ratings",
			"user = {:user} && book = {:book}",
			"", 100, 0,
			map[string]any{"user": user.Id, "book": book.Id},
		)
		for _, r := range existing {
			_ = app.Delete(r)
		}

		// Create new ratings
		coll, err := app.FindCollectionByNameOrId("genre_ratings")
		if err != nil {
			return err
		}
		for _, r := range data.Ratings {
			rec := core.NewRecord(coll)
			rec.Set("user", user.Id)
			rec.Set("book", book.Id)
			rec.Set("genre", r.Genre)
			rec.Set("rating", r.Rating)
			_ = app.Save(rec)
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Genre ratings saved"})
	}
}
