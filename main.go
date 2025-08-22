package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/jmartaudio/gator/internal/config"
	"github.com/jmartaudio/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please supply a command")
		os.Exit(1)
	}
	var s state
	var cmd command
	cmds := commands{
		handlers: make(map[string]func(*state, command) error),
	}

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config: %v", err)
		os.Exit(1)
	}
	s.cfg = &cfg

	dbURL := s.cfg.DBURL

	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)
	s.db = dbQueries

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerGetUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerAddFeed)
	cmds.register("feeds", handlerShowFeeds)
	cmds.register("follow", handlerFollow)
	cmds.register("following", handlerFollows)

	cmd.Name = os.Args[1]
	cmd.Args = os.Args[2:]

	err = cmds.run(&s, cmd)
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
