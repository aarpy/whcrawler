package main

import (
	"time"

	"github.com/aarpy/wisehoot/crawler/agent"
)

const (
	devKey   = "c980e3546ef3099f8e9cc68f3cce3a62"
	dbServer = "localhost:1234"
)

func main() {
	_ = agent.NewPastebin(time.Minute, devKey)
}
