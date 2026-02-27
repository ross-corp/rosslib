package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/pocketbase/pocketbase/core"
)

// GetMyGoals handles GET /me/goals — list all goals for the current user.
func GetMyGoals(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		records, err := app.FindRecordsByFilter("reading_goals",
			"user = {:user}", "-year", 100, 0,
			map[string]any{"user": user.Id},
		)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var goals []map[string]any
		for _, r := range records {
			goals = append(goals, map[string]any{
				"id":     r.Id,
				"year":   r.GetInt("year"),
				"target": r.GetInt("target"),
			})
		}
		if goals == nil {
			goals = []map[string]any{}
		}

		return e.JSON(http.StatusOK, goals)
	}
}

// UpsertGoal handles PUT /me/goals/{year} — create or update a reading goal for a year.
func UpsertGoal(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		yearStr := e.Request.PathValue("year")
		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 1900 || year > 2200 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid year"})
		}

		data := struct {
			Target int `json:"target"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}
		if data.Target < 1 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Target must be at least 1"})
		}

		// Find existing goal for this user+year
		existing, _ := app.FindRecordsByFilter("reading_goals",
			"user = {:user} && year = {:year}",
			"", 1, 0,
			map[string]any{"user": user.Id, "year": year},
		)

		var rec *core.Record
		if len(existing) > 0 {
			rec = existing[0]
		} else {
			coll, err := app.FindCollectionByNameOrId("reading_goals")
			if err != nil {
				return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to find collection"})
			}
			rec = core.NewRecord(coll)
			rec.Set("user", user.Id)
			rec.Set("year", year)
		}
		rec.Set("target", data.Target)

		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save goal"})
		}

		recordActivity(app, user.Id, "goal_set", map[string]any{
			"metadata": fmt.Sprintf(`{"year":%d,"target":%d}`, year, data.Target),
		})

		return e.JSON(http.StatusOK, map[string]any{
			"id":     rec.Id,
			"year":   rec.GetInt("year"),
			"target": rec.GetInt("target"),
		})
	}
}

// GetMyGoalYear handles GET /me/goals/{year} — get goal + progress for a year.
func GetMyGoalYear(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		yearStr := e.Request.PathValue("year")
		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 1900 || year > 2200 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid year"})
		}

		goals, err := app.FindRecordsByFilter("reading_goals",
			"user = {:user} && year = {:year}",
			"", 1, 0,
			map[string]any{"user": user.Id, "year": year},
		)
		if err != nil || len(goals) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "No goal found for this year"})
		}
		goal := goals[0]

		progress := countFinishedBooksInYear(app, user.Id, year)

		return e.JSON(http.StatusOK, map[string]any{
			"id":       goal.Id,
			"year":     goal.GetInt("year"),
			"target":   goal.GetInt("target"),
			"progress": progress,
		})
	}
}

// GetUserGoalYear handles GET /users/{username}/goals/{year} — public endpoint, respects privacy.
func GetUserGoalYear(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		username := e.Request.PathValue("username")
		yearStr := e.Request.PathValue("year")
		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 1900 || year > 2200 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid year"})
		}

		users, err := app.FindRecordsByFilter("users",
			"username = {:username}", "", 1, 0,
			map[string]any{"username": username},
		)
		if err != nil || len(users) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "User not found"})
		}
		user := users[0]

		viewerID := ""
		if e.Auth != nil {
			viewerID = e.Auth.Id
		}
		if !canViewProfile(app, viewerID, user) {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Profile is private"})
		}

		goals, err := app.FindRecordsByFilter("reading_goals",
			"user = {:user} && year = {:year}",
			"", 1, 0,
			map[string]any{"user": user.Id, "year": year},
		)
		if err != nil || len(goals) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "No goal found for this year"})
		}
		goal := goals[0]

		progress := countFinishedBooksInYear(app, user.Id, year)

		return e.JSON(http.StatusOK, map[string]any{
			"id":       goal.Id,
			"year":     goal.GetInt("year"),
			"target":   goal.GetInt("target"),
			"progress": progress,
		})
	}
}

// countFinishedBooksInYear counts books with "finished" status and date_read in the given year.
func countFinishedBooksInYear(app core.App, userID string, year int) int {
	startDate := fmt.Sprintf("%d-01-01 00:00:00", year)
	endDate := fmt.Sprintf("%d-01-01 00:00:00", year+1)

	type countResult struct {
		Count int `db:"count"`
	}
	var result countResult
	_ = app.DB().NewQuery(`
		SELECT COUNT(DISTINCT ub.book) as count
		FROM user_books ub
		JOIN book_tag_values btv ON btv.user = ub.user AND btv.book = ub.book
		JOIN tag_values tv ON btv.tag_value = tv.id
		WHERE ub.user = {:user}
		  AND tv.slug = 'finished'
		  AND ub.date_read >= {:start}
		  AND ub.date_read < {:end}
	`).Bind(map[string]any{
		"user":  userID,
		"start": startDate,
		"end":   endDate,
	}).One(&result)

	return result.Count
}
