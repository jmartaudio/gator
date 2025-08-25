package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jmartaudio/gator/internal/config"
	"github.com/jmartaudio/gator/internal/database"
	"github.com/lib/pq"
	"github.com/microcosm-cc/bluemonday"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	Name string
	Args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	err := c.handlers[cmd.Name](s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return errors.New("Must provide a username\n")
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

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return errors.New("Must provide a Name")
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
		fmt.Println("An error adding User to db")
		os.Exit(1)
	}
	s.cfg.SetUser(name)

	fmt.Println("Successfully created new User")
	printUser(user)

	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		fmt.Println("could not reset user db")
		os.Exit(1)
	}
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	if len(cmd.Args) > 0 {
		fmt.Println("this command takes no args")
	}
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println("Could not retirive users from db")
		os.Exit(1)
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

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		fmt.Println("Must provide a duration: 1s, 1m, 1h")
		os.Exit(1)
	}
	time_between_reqs, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time_between_reqs)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 2 {
		fmt.Println("Run command name url")
		os.Exit(1)
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

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		fmt.Println("Include URL argument")
		os.Exit(1)
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

func handlerFollows(s *state, cmd command, user database.User) error {
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
		fmt.Println("Please supply URL")
		os.Exit(1)
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

func printUser(user database.User) {
	fmt.Printf(" * ID:        %v\n", user.ID)
	fmt.Printf(" * Name:      %v\n", user.Name)
	fmt.Printf(" * CreatedAt: %v\n", user.CreatedAt)
}

func printfeed(feed database.Feed) {
	fmt.Printf(" * ID:        %v\n", feed.ID)
	fmt.Printf(" * Name:      %v\n", feed.Name)
	fmt.Printf(" * CreatedAt: %v\n", feed.CreatedAt)
	fmt.Printf(" * UpdatedAt: %v\n", feed.UpdatedAt)
	fmt.Printf(" * URL: %v\n", feed.Url)
	fmt.Printf(" * UserID: %v\n", feed.UserID)
}

func printPost(post database.Post) {
	p := bluemonday.StrictPolicy()
	fmt.Printf(" * Title:        %v\n", p.Sanitize(post.Title.String))
	fmt.Printf(" * Published at: %v\n", p.Sanitize(post.PublishedAt.String))
	fmt.Printf(" * URL:          %v\n", p.Sanitize(post.Url))
	fmt.Printf(" * Description:  %v\n", p.Sanitize(post.Description.String))
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("User-Agent", "gator")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, err
	}

	var rss RSSFeed
	if err := xml.Unmarshal(data, &rss); err != nil {
		fmt.Printf("Error unmarshalling XML: %v\n", err)
	}

	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)
	for _, item := range rss.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}

	return &rss, nil
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
