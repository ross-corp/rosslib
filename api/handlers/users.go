package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// SearchUsers handles GET /users?q=...&sort=newest|books|followers
func SearchUsers(app core.App) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		q := e.Request.URL.Query().Get("q")
		sort := e.Request.URL.Query().Get("sort")
		page, _ := strconv.Atoi(e.Request.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		perPage := 20
		offset := (page - 1) * perPage

		// For "books" and "followers" sorts we need subquery ordering,
		// so use raw SQL. For "newest" (default), use FindRecordsByFilter.
		if sort == "books" || sort == "followers" {
			return searchUsersSorted(app, e, q, sort, perPage, offset)
		}

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

		// Filter out blocked users if viewer is authenticated
		viewerID := ""
		if e.Auth != nil {
			viewerID = e.Auth.Id
		}

		var results []map[string]any
		for _, r := range records {
			if viewerID != "" && r.Id != viewerID && isBlockedEitherDirection(app, viewerID, r.Id) {
				continue
			}
			var avatarURL *string
			if av := r.GetString("avatar"); av != "" {
				url := "/api/files/" + r.Collection().Id + "/" + r.Id + "/" + av
				avatarURL = &url
			}
			results = append(results, map[string]any{
				"user_id":      r.Id,
				"username":     r.GetString("username"),
				"display_name": r.GetString("display_name"),
				"avatar_url":   avatarURL,
			})
		}
		if results == nil {
			results = []map[string]any{}
		}

		return e.JSON(http.StatusOK, results)
	}
}

func searchUsersSorted(app core.App, e *core.RequestEvent, q, sort string, perPage, offset int) error {
	usersColl, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return e.JSON(http.StatusOK, []any{})
	}
	collID := usersColl.Id

	var orderBy string
	var joinClause string
	switch sort {
	case "books":
		joinClause = "LEFT JOIN user_books ub ON ub.\"user\" = u.id"
		orderBy = "COUNT(ub.id) DESC, u.created DESC"
	case "followers":
		joinClause = "LEFT JOIN follows f ON f.followee = u.id AND f.status = 'active'"
		orderBy = "COUNT(f.id) DESC, u.created DESC"
	}

	whereClause := "1=1"
	params := map[string]any{"limit": perPage, "offset": offset}
	if q != "" {
		whereClause = "(u.username LIKE {:q} OR u.display_name LIKE {:q})"
		params["q"] = "%" + q + "%"
	}

	type userRow struct {
		ID          string  `db:"id"`
		Username    string  `db:"username"`
		DisplayName *string `db:"display_name"`
		Avatar      *string `db:"avatar"`
	}

	var rows []userRow
	err = app.DB().NewQuery(`
		SELECT u.id, u.username, u.display_name, u.avatar
		FROM users u
		` + joinClause + `
		WHERE ` + whereClause + `
		GROUP BY u.id
		ORDER BY ` + orderBy + `
		LIMIT {:limit} OFFSET {:offset}
	`).Bind(params).All(&rows)
	if err != nil {
		return e.JSON(http.StatusOK, []any{})
	}

	results := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		var avatarURL *string
		if r.Avatar != nil && *r.Avatar != "" {
			url := "/api/files/" + collID + "/" + r.ID + "/" + *r.Avatar
			avatarURL = &url
		}
		results = append(results, map[string]any{
			"user_id":      r.ID,
			"username":     r.Username,
			"display_name": r.DisplayName,
			"avatar_url":   avatarURL,
		})
	}

	return e.JSON(http.StatusOK, results)
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

		// Check if blocked in either direction
		blockedByViewer := false
		blockedByTarget := false
		if viewerID != "" && viewerID != user.Id {
			blockedByViewer = isBlocked(app, viewerID, user.Id)
			blockedByTarget = isBlocked(app, user.Id, viewerID)
		}
		isBlockedEither := blockedByViewer || blockedByTarget

		isRestricted := isBlockedEither || (user.GetBool("is_private") && !canViewProfile(app, viewerID, user))

		// Compute stats
		followStatus := "none"
		isFollowing := false
		if viewerID != "" && viewerID != user.Id && !isBlockedEither {
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

		var friendsCount countResult
		_ = app.DB().NewQuery(`
			SELECT COUNT(*) as count
			FROM follows f1
			JOIN follows f2 ON f1.follower = f2.followee AND f1.followee = f2.follower
			WHERE f1.follower = {:id} AND f1.status = 'active' AND f2.status = 'active'
		`).Bind(map[string]any{"id": user.Id}).One(&friendsCount)

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

		// Count books finished this year
		var booksThisYear countResult
		startOfYear := fmt.Sprintf("%d-01-01 00:00:00", time.Now().Year())
		_ = app.DB().NewQuery(`
			SELECT COUNT(DISTINCT btv.book) as count
			FROM book_tag_values btv
			JOIN tag_values tv ON btv.tag_value = tv.id
			JOIN user_books ub ON ub.user = btv.user AND ub.book = btv.book
			WHERE btv.user = {:id} AND tv.slug = 'finished'
			AND ub.date_read >= {:startOfYear}
		`).Bind(map[string]any{"id": user.Id, "startOfYear": startOfYear}).One(&booksThisYear)

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
			"friends_count":   friendsCount.Count,
			"books_read":      booksRead.Count,
			"reviews_count":   reviewsCount.Count,
			"books_this_year": booksThisYear.Count,
			"average_rating":  avgRating.Avg,
			"is_restricted":   isRestricted,
			"is_blocked":      blockedByViewer,
		}

		return e.JSON(http.StatusOK, result)
	}
}

// GetUserStats handles GET /users/{username}/stats
func GetUserStats(app core.App) func(e *core.RequestEvent) error {
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

		uid := user.Id
		currentYear := time.Now().Year()

		// Books by year (from date_read on user_books for finished books)
		type yearCount struct {
			Year  int `db:"year" json:"year"`
			Count int `db:"count" json:"count"`
		}
		var booksByYear []yearCount
		_ = app.DB().NewQuery(`
			SELECT CAST(strftime('%Y', ub.date_read) AS INTEGER) as year,
			       COUNT(*) as count
			FROM user_books ub
			JOIN book_tag_values btv ON btv.user = ub.user AND btv.book = ub.book
			JOIN tag_values tv ON btv.tag_value = tv.id
			WHERE ub.user = {:uid}
			  AND tv.slug = 'finished'
			  AND ub.date_read IS NOT NULL AND ub.date_read != ''
			GROUP BY year
			ORDER BY year DESC
		`).Bind(map[string]any{"uid": uid}).All(&booksByYear)
		if booksByYear == nil {
			booksByYear = []yearCount{}
		}

		// Books by month (current year)
		type monthCount struct {
			Year  int `db:"year" json:"year"`
			Month int `db:"month" json:"month"`
			Count int `db:"count" json:"count"`
		}
		var booksByMonth []monthCount
		_ = app.DB().NewQuery(`
			SELECT CAST(strftime('%Y', ub.date_read) AS INTEGER) as year,
			       CAST(strftime('%m', ub.date_read) AS INTEGER) as month,
			       COUNT(*) as count
			FROM user_books ub
			JOIN book_tag_values btv ON btv.user = ub.user AND btv.book = ub.book
			JOIN tag_values tv ON btv.tag_value = tv.id
			WHERE ub.user = {:uid}
			  AND tv.slug = 'finished'
			  AND ub.date_read IS NOT NULL AND ub.date_read != ''
			  AND strftime('%Y', ub.date_read) = {:year}
			GROUP BY month
			ORDER BY month
		`).Bind(map[string]any{"uid": uid, "year": strconv.Itoa(currentYear)}).All(&booksByMonth)
		if booksByMonth == nil {
			booksByMonth = []monthCount{}
		}

		// Average rating
		type avgResult struct {
			Avg *float64 `db:"avg"`
		}
		var avgRating avgResult
		_ = app.DB().NewQuery(`
			SELECT AVG(rating) as avg FROM user_books
			WHERE user = {:uid} AND rating > 0
		`).Bind(map[string]any{"uid": uid}).One(&avgRating)

		// Rating distribution (1-5)
		type ratingBucket struct {
			Rating int `db:"rating" json:"rating"`
			Count  int `db:"count" json:"count"`
		}
		var ratingDist []ratingBucket
		_ = app.DB().NewQuery(`
			SELECT CAST(rating AS INTEGER) as rating, COUNT(*) as count
			FROM user_books
			WHERE user = {:uid} AND rating > 0
			GROUP BY CAST(rating AS INTEGER)
			ORDER BY rating
		`).Bind(map[string]any{"uid": uid}).All(&ratingDist)

		// Build full distribution map (1-5)
		distMap := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
		for _, b := range ratingDist {
			if b.Rating >= 1 && b.Rating <= 5 {
				distMap[b.Rating] = b.Count
			}
		}
		ratingDistribution := make([]map[string]int, 5)
		for i := 1; i <= 5; i++ {
			ratingDistribution[i-1] = map[string]int{"rating": i, "count": distMap[i]}
		}

		// Total books
		type countResult struct {
			Count int `db:"count"`
		}
		var totalBooks countResult
		_ = app.DB().NewQuery(`
			SELECT COUNT(*) as count FROM user_books WHERE user = {:uid}
		`).Bind(map[string]any{"uid": uid}).One(&totalBooks)

		// Total reviews
		var totalReviews countResult
		_ = app.DB().NewQuery(`
			SELECT COUNT(*) as count FROM user_books
			WHERE user = {:uid} AND review_text IS NOT NULL AND review_text != ''
		`).Bind(map[string]any{"uid": uid}).One(&totalReviews)

		// Total pages read (sum page_count for finished books)
		type sumResult struct {
			Total *int `db:"total"`
		}
		var totalPages sumResult
		_ = app.DB().NewQuery(`
			SELECT COALESCE(SUM(b.page_count), 0) as total
			FROM user_books ub
			JOIN books b ON ub.book = b.id
			JOIN book_tag_values btv ON btv.user = ub.user AND btv.book = ub.book
			JOIN tag_values tv ON btv.tag_value = tv.id
			WHERE ub.user = {:uid}
			  AND tv.slug = 'finished'
			  AND b.page_count IS NOT NULL AND b.page_count > 0
		`).Bind(map[string]any{"uid": uid}).One(&totalPages)

		totalPagesRead := 0
		if totalPages.Total != nil {
			totalPagesRead = *totalPages.Total
		}

		return e.JSON(http.StatusOK, map[string]any{
			"books_by_year":       booksByYear,
			"books_by_month":      booksByMonth,
			"average_rating":      avgRating.Avg,
			"rating_distribution": ratingDistribution,
			"total_books":         totalBooks.Count,
			"total_reviews":       totalReviews.Count,
			"total_pages_read":    totalPagesRead,
		})
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
			LikeCount  int      `db:"like_count" json:"like_count"`
		}

		var reviews []reviewRow
		err = app.DB().NewQuery(`
			SELECT ub.rating, ub.review_text, ub.spoiler, ub.date_read, ub.date_added as date_added,
				   b.open_library_id, b.title,
				   COALESCE(NULLIF(ub.selected_edition_cover_url, ''), b.cover_url) as cover_url,
				   COALESCE((SELECT COUNT(*) FROM review_likes rl WHERE rl.book = ub.book AND rl.review_user = ub.user), 0) as like_count
			FROM user_books ub
			JOIN books b ON ub.book = b.id
			WHERE ub.user = {:user} AND ub.review_text != '' AND ub.review_text IS NOT NULL
			ORDER BY ub.date_added DESC
			LIMIT {:limit}
		`).Bind(map[string]any{"user": user.Id, "limit": limit}).All(&reviews)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		return e.JSON(http.StatusOK, reviews)
	}
}

// GetUserFollowers handles GET /users/{username}/followers
func GetUserFollowers(app core.App) func(e *core.RequestEvent) error {
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
		collID := user.Collection().Id

		viewerID := ""
		if e.Auth != nil {
			viewerID = e.Auth.Id
		}
		if user.GetBool("is_private") && !canViewProfile(app, viewerID, user) {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Profile is private"})
		}

		page, _ := strconv.Atoi(e.Request.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(e.Request.URL.Query().Get("limit"))
		if limit <= 0 || limit > 50 {
			limit = 50
		}
		offset := (page - 1) * limit

		type followerRow struct {
			UserID      string  `db:"user_id"`
			Username    string  `db:"username"`
			DisplayName *string `db:"display_name"`
			Avatar      string  `db:"avatar"`
		}

		var rows []followerRow
		err = app.DB().NewQuery(`
			SELECT u.id as user_id, u.username, u.display_name, u.avatar
			FROM follows f
			JOIN users u ON f.follower = u.id
			WHERE f.followee = {:user} AND f.status = 'active'
			ORDER BY f.created DESC
			LIMIT {:limit} OFFSET {:offset}
		`).Bind(map[string]any{"user": user.Id, "limit": limit, "offset": offset}).All(&rows)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		results := make([]map[string]any, 0, len(rows))
		for _, r := range rows {
			var avatarURL *string
			if r.Avatar != "" {
				u := "/api/files/" + collID + "/" + r.UserID + "/" + r.Avatar
				avatarURL = &u
			}
			results = append(results, map[string]any{
				"user_id":      r.UserID,
				"username":     r.Username,
				"display_name": r.DisplayName,
				"avatar_url":   avatarURL,
			})
		}

		return e.JSON(http.StatusOK, results)
	}
}

// GetUserFollowing handles GET /users/{username}/following
func GetUserFollowing(app core.App) func(e *core.RequestEvent) error {
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
		collID := user.Collection().Id

		viewerID := ""
		if e.Auth != nil {
			viewerID = e.Auth.Id
		}
		if user.GetBool("is_private") && !canViewProfile(app, viewerID, user) {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Profile is private"})
		}

		page, _ := strconv.Atoi(e.Request.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(e.Request.URL.Query().Get("limit"))
		if limit <= 0 || limit > 50 {
			limit = 50
		}
		offset := (page - 1) * limit

		type followingRow struct {
			UserID      string  `db:"user_id"`
			Username    string  `db:"username"`
			DisplayName *string `db:"display_name"`
			Avatar      string  `db:"avatar"`
		}

		var rows []followingRow
		err = app.DB().NewQuery(`
			SELECT u.id as user_id, u.username, u.display_name, u.avatar
			FROM follows f
			JOIN users u ON f.followee = u.id
			WHERE f.follower = {:user} AND f.status = 'active'
			ORDER BY f.created DESC
			LIMIT {:limit} OFFSET {:offset}
		`).Bind(map[string]any{"user": user.Id, "limit": limit, "offset": offset}).All(&rows)
		if err != nil {
			return e.JSON(http.StatusOK, []any{})
		}

		results := make([]map[string]any, 0, len(rows))
		for _, r := range rows {
			var avatarURL *string
			if r.Avatar != "" {
				u := "/api/files/" + collID + "/" + r.UserID + "/" + r.Avatar
				avatarURL = &u
			}
			results = append(results, map[string]any{
				"user_id":      r.UserID,
				"username":     r.Username,
				"display_name": r.DisplayName,
				"avatar_url":   avatarURL,
			})
		}

		return e.JSON(http.StatusOK, results)
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

		if data.DisplayName != nil && len(*data.DisplayName) > 100 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Display name must be 100 characters or fewer"})
		}
		if data.Bio != nil && len(*data.Bio) > 2000 {
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Bio must be 2000 characters or fewer"})
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

		// Check if blocked in either direction
		if isBlockedEitherDirection(app, user.Id, target.Id) {
			return e.JSON(http.StatusForbidden, map[string]any{"error": "Cannot follow this user"})
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

		recordActivity(app, user.Id, "followed_user", map[string]any{"target_user": target.Id})

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
