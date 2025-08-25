package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmartaudio/gator/internal/database"
)

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("Must provide a Name")
	}

	name := cmd.Args[0]

	user, err := s.db.CreateUser(context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Name:      name,
		})
	if err != nil {
		return fmt.Errorf("An error adding User to db")
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("Successfully created new User")
	printUser(user)
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("Must provide a username\n")
	}

	name := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("couldn't find user: %w\n", err)
	}

	err = s.cfg.SetUser(cmd.Args[0])
	if err != nil {
		return err
	}

	fmt.Printf("The user has been set to: %s\n", s.cfg.CurrentUserName)

	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Could not retirive users: %w from db", err)
	}
	for _, user := range users {
		if user == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user)
		} else {
			fmt.Printf("* %s\n", user)
		}
	}
	return nil
}

func printUser(user database.User) {
	fmt.Printf(" * ID:        %v\n", user.ID)
	fmt.Printf(" * Name:      %v\n", user.Name)
	fmt.Printf(" * CreatedAt: %v\n", user.CreatedAt)
}
