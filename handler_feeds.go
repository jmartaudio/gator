package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmartaudio/gator/internal/database"
)

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("Command: %s <URL>", cmd.Name)
	}

	url := cmd.Args[0]

	feed, err := s.db.GetFeedsByUrl(context.Background(), url)
	if err != nil {
		return err
	}

	follow, err := s.db.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			UserID:    user.ID,
			FeedID:    feed.ID,
		},
	)

	fmt.Printf("%s is following %s\n", follow.UserName, follow.FeedName)

	return nil
}

func handlerListFollows(s *state, cmd command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)

	if err != nil {
		return err
	}

	if len(follows) < 1 {
		fmt.Println("You are not following anything")
	} else {
		fmt.Printf("%s is following:\n", user.Name)
		for _, follow := range follows {
			fmt.Println(follow.FeedName)
		}
	}

	return nil
}

func handlerUnFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("Commnad: %s <URL>", cmd.Name)
	}

	url := cmd.Args[0]

	feed, err := s.db.GetFeedsByUrl(context.Background(), url)
	if err != nil {
		return err
	}

	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return nil
	}

	fmt.Printf("%s is unfollowing %s\n", user.Name, feed.Name)

	return nil
}
