package main

import (
	"time"

	"github.com/aarpy/whcrawler/agent"
)

const (
	devKey   = ""
	dbServer = "localhost:1234"
)

func main() {
	_ = agent.NewPastebin(time.Minute, devKey)
}
