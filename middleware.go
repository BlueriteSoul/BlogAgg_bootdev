package main

import (
	"fmt"
)

middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	fmt.Printf("fkjdbhjd")
	return nil
}