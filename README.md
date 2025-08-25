# Gator

Gator is an rss aggrigator built in Go
It is a project from Boot.Dev

I've been coding for like 3 months so proceed with caution

This project requires Postgres and Go

- Go to build it [Go toolchain](https://golang.org/dl/)
- Postgres to store the data

## Config file

- create a file in your home dir called .gatorconfig.json

```
{
  "db_url": "postgres://username:@localhost:5432/database?sslmode=disable"
}
```

Your username for gator will be added to this file by the program

## Installation

To install

- clone the repo
- cd into gator
- go install to install in your Go dir
- or go build to run it from the gator dir

The commands

- gator register <your_name> -- Adds a user
- gator login <your_name> -- Changes the user
- gator addfeed <url> -- Add a feed to the db
- gator follow <name> <url> -- Follow a feed that already exists in the database
- gator unfollow <url> -- Unfollow a feed that already exists in the database
- gator browse <limit> Browse x number of recent postes example "browse 10"
- gator agg <time_interval> -- Example "agg 5m" to fetch new posts every 5m

leave agg running like a daemon in another terminal
for continuous fetching
