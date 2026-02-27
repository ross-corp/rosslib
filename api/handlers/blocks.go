package handlers

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

// isBlocked checks whether blockerID has blocked blockedID.
func isBlocked(app core.App, blockerID, blockedID string) bool {
	records, err := app.FindRecordsByFilter("blocks",
		"blocker = {:blocker} && blocked = {:blocked}",
		"", 1, 0,
		map[string]any{"blocker": blockerID, "blocked": blockedID},
	)
	return err == nil && len(records) > 0
}

// isBlockedEitherDirection checks whether a block exists in either direction.
func isBlockedEitherDirection(app core.App, userA, userB string) bool {
	return isBlocked(app, userA, userB) || isBlocked(app, userB, userA)
}

// BlockUser handles POST /users/{username}/block
func BlockUser(app core.App) func(e *core.RequestEvent) error {
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
			return e.JSON(http.StatusBadRequest, map[string]any{"error": "Cannot block yourself"})
		}

		// Check if already blocked
		if isBlocked(app, user.Id, target.Id) {
			return e.JSON(http.StatusOK, map[string]any{"blocked": true})
		}

		// Create block record
		blocksColl, err := app.FindCollectionByNameOrId("blocks")
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Internal error"})
		}
		rec := core.NewRecord(blocksColl)
		rec.Set("blocker", user.Id)
		rec.Set("blocked", target.Id)
		if err := app.Save(rec); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to block user"})
		}

		// Remove existing follows in both directions
		removeFollows(app, user.Id, target.Id)
		removeFollows(app, target.Id, user.Id)

		return e.JSON(http.StatusOK, map[string]any{"blocked": true})
	}
}

// removeFollows deletes follow records from follower → followee.
func removeFollows(app core.App, followerID, followeeID string) {
	follows, err := app.FindRecordsByFilter("follows",
		"follower = {:f} && followee = {:t}",
		"", 1, 0,
		map[string]any{"f": followerID, "t": followeeID},
	)
	if err == nil && len(follows) > 0 {
		_ = app.Delete(follows[0])
	}
}

// UnblockUser handles DELETE /users/{username}/block
func UnblockUser(app core.App) func(e *core.RequestEvent) error {
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

		blocks, err := app.FindRecordsByFilter("blocks",
			"blocker = {:blocker} && blocked = {:blocked}",
			"", 1, 0,
			map[string]any{"blocker": user.Id, "blocked": target.Id},
		)
		if err != nil || len(blocks) == 0 {
			return e.JSON(http.StatusOK, map[string]any{"blocked": false})
		}

		if err := app.Delete(blocks[0]); err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to unblock"})
		}

		return e.JSON(http.StatusOK, map[string]any{"blocked": false})
	}
}

// CheckBlock handles GET /users/{username}/block
func CheckBlock(app core.App) func(e *core.RequestEvent) error {
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

		return e.JSON(http.StatusOK, map[string]any{
			"blocked": isBlocked(app, user.Id, target.Id),
		})
	}
}
