package main

import (
	"context"
	"fmt"

	"github.com/BlueriteSoul/BlogAgg_bootdev/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		// Get the user from the database
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			// Return an error if fetching the user fails
			return fmt.Errorf("couldn't get user, error: %w", err)
		}

		// Optionally, log the user details or perform other actions before invoking the handler
		fmt.Printf("Logged in user: %s\n", user.Name)

		// Call the original handler with the user
		return handler(s, cmd, user) // Calls the original handler
	}
}
