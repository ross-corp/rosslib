package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	UserIDKey      = "user_id"
	UsernameKey    = "username"
	IsModeratorKey = "is_moderator"
)

func Auth(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := parseToken(c, jwtSecret)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set(UserIDKey, claims["sub"].(string))
		c.Set(UsernameKey, claims["username"].(string))
		if mod, ok := claims["is_moderator"].(bool); ok {
			c.Set(IsModeratorKey, mod)
		}
		c.Next()
	}
}

func OptionalAuth(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		if claims, ok := parseToken(c, jwtSecret); ok {
			c.Set(UserIDKey, claims["sub"].(string))
			c.Set(UsernameKey, claims["username"].(string))
			if mod, ok := claims["is_moderator"].(bool); ok {
				c.Set(IsModeratorKey, mod)
			}
		}
		c.Next()
	}
}

func RequireModerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		isMod, _ := c.Get(IsModeratorKey)
		if isMod != true {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "moderator access required"})
			return
		}
		c.Next()
	}
}

func parseToken(c *gin.Context, secret []byte) (jwt.MapClaims, bool) {
	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return nil, false
	}
	tokenStr := strings.TrimPrefix(h, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	return claims, ok
}
