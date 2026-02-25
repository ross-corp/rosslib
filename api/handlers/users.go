package handlers

import (
	"net/http"
	"strconv"

	"github.com/pocketbase/pocketbase/core"
)

// SearchUsers handles GET /users?q=...
func SearchUsers(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		q := e.Request.URL.Query().Get("q")
		page, _ := strconv.Atoi(e.Request.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		perPage := 20
		offset := (page - 1) * perPage

		filter := "1=1"
		params := map[string]any{}
		if q != "" {
			filter = "username LIKE {:q} || display_name LIKE {:q}"
			params["q"] = "%" + q + "%"
		}

		records, err := app.FindRecordsByFilter("users", filter, "-created", perPage, offset, params)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		var results []map[string]any
		for _, r := range records {
			results = append(results, map[string]any{
				"user_id":      r.Id,
				"username":     r.GetString("username"),
				"display_name": r.GetString("display_name"),
			})
		}
		if results == nil {
			results = []map[string]any{}
		}

		return e.JSON(http.StatusOK, results)
	}
}

// GetProfile handles GET /users/{username}
func GetProfile(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		username := e.Request.PathValue("username")

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

		isRestricted := user.GetBool("is_private") && !canViewProfile(app, viewerID, user)

		// Compute stats
		followStatus := "none"
		isFollowing := false
		if viewerID != "" && viewerID != user.Id {
			follows, err := app.FindRecordsByFilter("follows",
				"follower = {:viewer} && followee = {:target}",
				"", 1, 0,
				map[string]any{"viewer": viewerID, "target": user.Id},
			)
			if err == nil && len(follows) > 0 {
				followStatus = follows[0].GetString("status")
				isFollowing = followStatus == "active"
			}
		}

		// Count followers/following
		type countResult struct {
			Count int `db:"count"`
		}
		var followersCount, followingCount countResult
		_ = app.DB().NewQuery("SELECT COUNT(*) as count FROM follows WHERE followee = {:id} AND status = 'active'").
			Bind(map[string]any{"id": user.Id}).One(&followersCount)
		_ = app.DB().NewQuery("SELECT COUNT(*) as count FROM follows WHERE follower = {:id} AND status = 'active'").
			Bind(map[string]any{"id": user.Id}).One(&followingCount)

		// Count books read & reviews
		var booksRead, reviewsCount countResult
		_ = app.DB().NewQuery(`
			SELECT COUNT(DISTINCT btv.book) as count
			FROM book_tag_values btv
			JOIN tag_values tv ON btv.tag_value = tv.id
			WHERE btv.user = {:id} AND tv.slug = 'finished'
		`).Bind(map[string]any{"id": user.Id}).One(&booksRead)
		_ = app.DB().NewQuery(`
			SELECT COUNT(*) as count FROM user_books
			WHERE user = {:id} AND review_text != '' AND review_text IS NOT NULL
		`).Bind(map[string]any{"id": user.Id}).One(&reviewsCount)

		// Average rating
		type avgResult struct {
			Avg *float64 `db:"avg"`
		}
		var avgRating avgResult
		_ = app.DB().NewQuery(`
			SELECT AVG(rating) as avg FROM user_books
			WHERE user = {:id} AND rating > 0
		`).Bind(map[string]any{"id": user.Id}).One(&avgRating)

		// Avatar URL
		var avatarURL *string
		if av := user.GetString("avatar"); av != "" {
			url := "/api/files/" + user.Collection().Id + "/" + user.Id + "/" + av
			avatarURL = &url
		}

		result := map[string]any{
			"user_id":         user.Id,
			"username":        user.GetString("username"),
			"display_name":    user.GetString("display_name"),
			"bio":             user.GetString("bio"),
			"avatar_url":      avatarURL,
			"is_private":      user.GetBool("is_private"),
			"member_since":    user.GetString("created"),
			"is_following":    isFollowing,
			"follow_status":   followStatus,
			"followers_count": followersCount.Count,
			"following_count": followingCount.Count,
			"friends_count":   0,
			"books_read":      booksRead.Count,
			"reviews_count":   reviewsCount.Count,
			"books_this_year": 0,
			"average_rating":  avgRating.Avg,
			"is_restricted":   isRestricted,
		}

		return e.JSON(http.StatusOK, result)
	}
}

// GetUserReviews handles GET /users/{username}/reviews
func GetUserReviews(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		username := e.Request.PathValue("username")

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

		limit, _ := strconv.Atoi(e.Request.URL.Query().Get("limit"))
		if limit <= 0 || limit > 50 {
			limit = 20
		}

		type reviewRow struct {
			Rating     *float64 `db:"rating" json:"rating"`
			ReviewText string   `db:"review_text" json:"review_text"`
			Spoiler    bool     `db:"spoiler" json:"spoiler"`
			DateRead   *string  `db:"date_read" json:"date_read"`
			DateAdded  string   `db:"date_added" json:"date_added"`
			BookOLID   string   `db:"open_library_id" json:"open_library_id"`
			BookTitle  string   `db:"title" json:"title"`
			CoverURL   *string  `db:"cover_url" json:"cover_url"`
		}

		var reviews []reviewRow
		err = app.DB().NewQuery(`
			SELECT ub.rating, ub.review_text, ub.spoiler, ub.date_read, ub.created as date_added,
				   b.open_library_id, b.title, b.cover_url
			FROM user_books ub
			JOIN books b ON ub.book = b.id
			WHERE ub.user = {:user} AND ub.review_text != '' AND ub.review_text IS NOT NULL
			ORDER BY ub.created DESC
			LIMIT {:limit}
		`).Bind(map[string]any{"user": user.Id, "limit": limit}).All(&reviews)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		return e.JSON(http.StatusOK, reviews)
	}
}

// UpdateProfile handles PATCH /users/me
func UpdateProfile(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		data := struct {
			DisplayName *string `json:"display_name"`
			Bio         *string `json:"bio"`
			IsPrivate   *bool   `json:"is_private"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid request body"})
		}

		if data.DisplayName != nil {
			user.Set("display_name", *data.DisplayName)
		}
		if data.Bio != nil {
			user.Set("bio", *data.Bio)
		}
		if data.IsPrivate != nil {
			user.Set("is_private", *data.IsPrivate)
		}

		if err := app.Save(user); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return e.JSON(http.StatusOK, map[string]any{
			"user_id":      user.Id,
			"username":     user.GetString("username"),
			"display_name": user.GetString("display_name"),
			"bio":          user.GetString("bio"),
			"is_private":   user.GetBool("is_private"),
		})
	}
}

// UploadAvatar handles POST /me/avatar
func UploadAvatar(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		file, header, err := e.Request.FormFile("avatar")
		if err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "No avatar file provided"})
		}
		defer file.Close()

		// Use PocketBase's filesystem to handle file upload
		f, err := e.FindUploadedFiles("avatar")
		if err != nil || len(f) == 0 {
			_ = header // suppress unused warning
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Failed to process uploaded file"})
		}

		user.Set("avatar", f[0])
		if err := app.Save(user); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to save avatar"})
		}

		avatarURL := "/api/files/" + user.Collection().Id + "/" + user.Id + "/" + user.GetString("avatar")
		return e.JSON(http.StatusOK, map[string]any{
			"avatar_url": avatarURL,
		})
	}
}

// FollowUser handles POST /users/{username}/follow
func FollowUser(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		username := e.Request.PathValue("username")
		targets, err := app.FindRecordsByFilter("users",
			"username = {:username}", "", 1, 0,
			map[string]any{"username": username},
		)
		if err != nil || len(targets) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "User not found"})
		}
		target := targets[0]

		if target.Id == user.Id {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Cannot follow yourself"})
		}

		// Check if already following
		existing, _ := app.FindRecordsByFilter("follows",
			"follower = {:follower} && followee = {:followee}",
			"", 1, 0,
			map[string]any{"follower": user.Id, "followee": target.Id},
		)
		if len(existing) > 0 {
			return e.JSON(http.StatusOK, map[string]any{
				"status": existing[0].GetString("status"),
			})
		}

		status := "active"
		if target.GetBool("is_private") {
			status = "pending"
		}

		followsColl, err := app.FindCollectionByNameOrId("follows")
		if err != nil {
			return err
		}
		rec := core.NewRecord(followsColl)
		rec.Set("follower", user.Id)
		rec.Set("followee", target.Id)
		rec.Set("status", status)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		recordActivity(app, user.Id, "follow", map[string]any{"target_user": target.Id})

		return e.JSON(http.StatusOK, map[string]any{"status": status})
	}
}

// UnfollowUser handles DELETE /users/{username}/follow
func UnfollowUser(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		username := e.Request.PathValue("username")
		targets, err := app.FindRecordsByFilter("users",
			"username = {:username}", "", 1, 0,
			map[string]any{"username": username},
		)
		if err != nil || len(targets) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "User not found"})
		}
		target := targets[0]

		follows, err := app.FindRecordsByFilter("follows",
			"follower = {:follower} && followee = {:followee}",
			"", 1, 0,
			map[string]any{"follower": user.Id, "followee": target.Id},
		)
		if err != nil || len(follows) == 0 {
			return e.JSON(http.StatusOK, map[string]any{"message": "Not following"})
		}

		if err := app.Delete(follows[0]); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to unfollow"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Unfollowed"})
	}
}

// GetFollowRequests handles GET /me/follow-requests
func GetFollowRequests(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		type reqRow struct {
			FollowerID  string  `db:"follower" json:"user_id"`
			Username    string  `db:"username" json:"username"`
			DisplayName *string `db:"display_name" json:"display_name"`
		}

		var requests []reqRow
		err := app.DB().NewQuery(`
			SELECT f.follower, u.username, u.display_name
			FROM follows f
			JOIN users u ON f.follower = u.id
			WHERE f.followee = {:user} AND f.status = 'pending'
			ORDER BY f.created DESC
		`).Bind(map[string]any{"user": user.Id}).All(&requests)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		return e.JSON(http.StatusOK, requests)
	}
}

// AcceptFollowRequest handles POST /me/follow-requests/{userId}/accept
func AcceptFollowRequest(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		followerID := e.Request.PathValue("userId")
		follows, err := app.FindRecordsByFilter("follows",
			"follower = {:follower} && followee = {:followee} && status = 'pending'",
			"", 1, 0,
			map[string]any{"follower": followerID, "followee": user.Id},
		)
		if err != nil || len(follows) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Follow request not found"})
		}

		follows[0].Set("status", "active")
		if err := app.Save(follows[0]); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to accept"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Follow request accepted"})
	}
}

// RejectFollowRequest handles DELETE /me/follow-requests/{userId}/reject
func RejectFollowRequest(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		user := e.Auth
		if user == nil {
			return e.JSON(http.StatusUnauthorized, map[string]any{"error": "Authentication required"})
		}

		followerID := e.Request.PathValue("userId")
		follows, err := app.FindRecordsByFilter("follows",
			"follower = {:follower} && followee = {:followee} && status = 'pending'",
			"", 1, 0,
			map[string]any{"follower": followerID, "followee": user.Id},
		)
		if err != nil || len(follows) == 0 {
			return e.JSON(http.StatusNotFound, map[string]any{"error": "Follow request not found"})
		}

		if err := app.Delete(follows[0]); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to reject"})
		}

		return e.JSON(http.StatusOK, map[string]any{"message": "Follow request rejected"})
	}
}
