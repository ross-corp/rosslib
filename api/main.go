package main

import (
	"log"
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "github.com/tristansaldanha/rosslib/api/migrations"
)

func main() {
	app := pocketbase.New()

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: true,
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Custom /auth/login route to mimic existing API contract
		se.Router.POST("/auth/login", func(e *core.RequestEvent) error {
			data := struct {
				Email    string `json:"email"`
				Password string `json:"password"`
			}{}

			if err := e.BindBody(&data); err != nil {
				return err
			}

			record, err := app.FindAuthRecordByEmail("users", data.Email)
			if err != nil {
				return apis.NewBadRequestError("Invalid credentials", err)
			}

			if !record.ValidatePassword(data.Password) {
				return apis.NewBadRequestError("Invalid credentials", nil)
			}

			token, err := record.NewAuthToken()
			if err != nil {
				return err
			}

			return e.JSON(http.StatusOK, map[string]any{
				"token":    token,
				"user_id":  record.Id,
				"username": record.GetString("username"),
			})
		})

		// Custom /auth/register route
		se.Router.POST("/auth/register", func(e *core.RequestEvent) error {
			data := struct {
				Username        string `json:"username"`
				Email           string `json:"email"`
				Password        string `json:"password"`
				PasswordConfirm string `json:"passwordConfirm"`
			}{}

			if err := e.BindBody(&data); err != nil {
				return err
			}

			collection, err := app.FindCollectionByNameOrId("users")
			if err != nil {
				return err
			}

			record := core.NewRecord(collection)
			record.Set("username", data.Username)
			record.Set("email", data.Email)
			record.SetPassword(data.Password)

			if err := app.Save(record); err != nil {
				return err
			}

			token, err := record.NewAuthToken()
			if err != nil {
				return err
			}

			return e.JSON(http.StatusOK, map[string]any{
				"token":    token,
				"user_id":  record.Id,
				"username": record.GetString("username"),
			})
		})

		// Custom /me/feed route (Mock implementation to support frontend)
		se.Router.GET("/me/feed", func(e *core.RequestEvent) error {
			// In a real implementation, we would fetch follows and recent items.
			// For now, return empty feed to allow the page to load.
			return e.JSON(http.StatusOK, map[string]any{
				"activities": []any{},
				"next_cursor": nil,
			})
		})

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
