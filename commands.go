package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jmartaudio/gator/internal/config"
	"github.com/jmartaudio/gator/internal/database"
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
	url := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(context.Background(), url)
	if err != nil {
		fmt.Println("Could not fetch feed")
		os.Exit(1)
	}
	fmt.Printf("Retrived XML: %v\n", feed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.Args) < 2 {
		fmt.Println("Run command name url")
		os.Exit(1)
	}
	user_name := s.cfg.CurrentUserName
	user, err := s.db.GetUser(context.Background(), user_name)
	if err != nil {
		return err
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
	_, err = s.db.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			UserID:    user.ID,
			FeedID:    feed.ID,
		},
	)
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

func handlerFollow(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		fmt.Println("Include URL argument")
		os.Exit(1)
	}
	url := cmd.Args[0]
	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}
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

func handlerFollows(s *state, cmd command) error {
	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return err
	}
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
