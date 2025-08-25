package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmartaudio/gator/internal/database"
)

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("Run command %s witn <name> <url>", cmd.Name)
	}
	name := cmd.Args[0]
	url := cmd.Args[1]

	feed, err := s.db.AddFeed(context.Background(),
		database.AddFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Name:      name,
			Url:       url,
			UserID:    user.ID,
		},
	)
	if err != nil {
		return err
	}
	_, err = s.db.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			UserID:    user.ID,
			FeedID:    feed.ID,
		},
	)
	if err != nil {
		return err
	}
	printfeed(feed)
	return nil
}

func handlerShowFeeds(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		user_id, err := s.db.GetUser(context.Background(), user)
		feedList, err := s.db.GetFeeds(context.Background(), user_id.ID)
		if err != nil {
			fmt.Println("Nothing to show")
			return err
		}
		for _, feed := range feedList {
			if len(feedList) < 1 {
				continue
			} else {
				fmt.Println(user)
				fmt.Printf("Name: %s URL: %s\n", feed.Name, feed.Url)
			}
		}
	}
	return nil
}

func printfeed(feed database.Feed) {
	fmt.Printf(" * ID:        %v\n", feed.ID)
	fmt.Printf(" * Name:      %v\n", feed.Name)
	fmt.Printf(" * CreatedAt: %v\n", feed.CreatedAt)
	fmt.Printf(" * UpdatedAt: %v\n", feed.UpdatedAt)
	fmt.Printf(" * URL: %v\n", feed.Url)
	fmt.Printf(" * UserID: %v\n", feed.UserID)
}
