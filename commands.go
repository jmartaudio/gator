package main

import "github.com/jmartaudio/gator/internal/config"

type State struct {
	State *config
}

type command struct {
	name     string
	argument []string
}
