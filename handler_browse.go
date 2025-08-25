package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jmartaudio/gator/internal/database"
	"github.com/microcosm-cc/bluemonday"
)

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit int32
	limit = 2
	if len(cmd.Args) > 0 {
		argInt, err := strconv.ParseInt(cmd.Args[0], 10, 32)
		if err != nil {
			fmt.Println("Provide limit as an integer")
		}
		limit = int32(argInt)
	}
	feeds, err := s.db.GetFeeds(context.Background(), user.ID)
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		posts, err := s.db.GetPostForUser(context.Background(),
			database.GetPostForUserParams{
				FeedID: feed.ID,
				Limit:  limit,
			},
		)
		if err != nil {
			return err
		}
		for _, post := range posts {
			printPost(post)
		}
	}

	return nil
}

func printPost(post database.Post) {
	p := bluemonday.StrictPolicy()
	fmt.Printf(" * Title:        %v\n", p.Sanitize(post.Title.String))
	fmt.Printf(" * Published at: %v\n", p.Sanitize(post.PublishedAt.String))
	fmt.Printf(" * URL:          %v\n", p.Sanitize(post.Url))
	fmt.Printf(" * Description:  %v\n", p.Sanitize(post.Description.String))
}
