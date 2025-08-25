package main

import (
	"context"
	"fmt"
)

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not reset user db %w", err)
	}

	fmt.Println("Database reset successfully!")
	return nil
}
