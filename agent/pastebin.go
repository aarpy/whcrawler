// Package agent wraps the basic functions of the Pastebin API and exposes a
// Go API.
// source: https://github.com/bearbin/go-paste/blob/master/pastebin/pastebin.go
package agent

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

type pastebin struct {
	lastCrawlTime time.Time
	crawlDuration time.Duration
	devKey        string
}

// NewPastebin function
func NewPastebin(duration time.Duration, devKey string) Agent {
	log.Info("New Pastbin agent:", duration)

	return &pastebin{time.Now(), duration, devKey}
}

// Put uploads text to Pastebin with optional title returning the ID or an error.
func (p pastebin) Put(text, title string) (id string, err error) {
	data := url.Values{}
	// Required values.
	data.Set("api_dev_key", p.devKey)
	data.Set("api_option", "paste") // Create a paste.
	data.Set("api_paste_code", text)
	// Optional values.
	data.Set("api_paste_name", title)      // The paste should have title "title".
	data.Set("api_paste_private", "0")     // Create a public paste.
	data.Set("api_paste_expire_date", "N") // The paste should never expire.

	resp, err := http.PostForm("http://pastebin.com/api/api_post.php", data)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New("Pastebin put failed")
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return p.StripURL(string(respBody)), nil
}

// Get returns the text inside the paste identified by ID.
func (p pastebin) Get(id string) (text string, err error) {
	resp, err := http.Get("http://pastebin.com/raw.php?i=" + id)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New("Pastebin get failed")
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(respBody), nil
}

// StripURL returns the paste ID from a pastebin URL.
func (p pastebin) StripURL(url string) string {
	return strings.Replace(url, "http://pastebin.com/", "", -1)
}

// WrapID returns the pastebin URL from a paste ID.
func (p pastebin) WrapID(id string) string {
	return "http://pastebin.com/" + id
}

func (p *pastebin) Run() (bool, error) {
	return true, nil
}
