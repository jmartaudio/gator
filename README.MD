# Gator

Gator is an rss aggrigator built in Go
It is a project from Boot.Dev

I've been coding for like 3 months so proceed with caution

This project requires Postgres and Go
- Go to build it
- Postgres to store the data

Config file
- create a file in your home dir called .gatorconfig.json

```
{
  "db_url": "postgres://example"
}
```
paste this and replace the example with your postgres db
should be like this:

protocol://username:password@host:port/database

Your username for gator will be added to this file by the program

To install
- clone the repo
- cd into gator
- go install to install in your Go dir
- or go build to run it from the gator dir

The commands
- gator register "your_name" -- adds a user
- gator login "your_name" -- changes the user
- gator follow "name" "url" -- follows an RSS feed
- gator unfollow "url" -- unfollows an RSS feed
- gator browse "limit" browse x number of recent postes example "browse 10"
- gator agg "time_interval" -- example "agg 5m" to fetch new posts every 5m

leave agg running like a daemon in another terminal
for continuous fetching
