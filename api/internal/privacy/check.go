package privacy

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CanViewProfile checks whether a viewer can see a user's private content.
// Returns the target user's ID, whether they are private, and whether the
// viewer is allowed to see their content (owner, active follower, or public).
func CanViewProfile(ctx context.Context, pool *pgxpool.Pool, username, currentUserID string) (userID string, isPrivate bool, canView bool) {
	err := pool.QueryRow(ctx,
		`SELECT id, is_private FROM users WHERE username = $1 AND deleted_at IS NULL`,
		username,
	).Scan(&userID, &isPrivate)
	if err != nil {
		return "", false, false
	}

	if !isPrivate {
		return userID, false, true
	}

	// Owner can always view their own profile
	if currentUserID == userID {
		return userID, true, true
	}

	// Check for active follow
	if currentUserID != "" {
		var isFollowing bool
		_ = pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = $2 AND status = 'active')`,
			currentUserID, userID,
		).Scan(&isFollowing)
		if isFollowing {
			return userID, true, true
		}
	}

	return userID, true, false
}
