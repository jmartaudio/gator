package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmartaudio/gator/internal/database"

	"github.com/lib/pq"
)

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("Command: %s provide a duration: 1s, 1m, 1h", cmd.Name)
	}
	timeBetweenReqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	log.Printf("Collecting feeds every %s...", timeBetweenReqs)

	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func scrapeFeeds(s *state) error {
	currentTime := time.Now()
	nullTime := sql.NullTime{Time: currentTime, Valid: true}

	next, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	err = s.db.MarkFeedFetched(context.Background(),
		database.MarkFeedFetchedParams{
			UpdatedAt:     currentTime,
			LastFetchedAt: nullTime,
			ID:            next.ID,
		},
	)
	feed, err := fetchFeed(context.Background(), next.Url)
	if err != nil {
		return err
	}
	for _, item := range feed.Channel.Item {
		post, err := s.db.CreatePost(context.Background(),
			database.CreatePostParams{
				ID:          uuid.New(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Title:       sql.NullString{String: item.Title, Valid: true},
				Url:         item.Link,
				Description: sql.NullString{String: item.Description, Valid: true},
				PublishedAt: sql.NullString{String: string(item.PubDate), Valid: true},
				FeedID:      next.ID,
			},
		)
		if err != nil {
			if err, ok := err.(*pq.Error); ok {
				if err.Code == "23505" {
					continue
				}
			} else {
				log.Println(err)
				return err
			}
		}
		fmt.Printf("New Post %s From %s\n", post.Title.String, next.Name)
	}

	return nil
}
