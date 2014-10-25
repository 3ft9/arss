package main

import (
	ring "container/ring"
	"fmt"
	rss "github.com/jteeuwen/go-pkg-rss"
	"os"
	"sync"
)

type State struct {
	sync.Mutex
	Feeds  map[string]*Feed
	Recent *ring.Ring
}

type Feed struct {
	Object       *rss.Feed
	CheckDueAt   int64
	ArticleCount int64
}

type RecentItem struct {
	Source string
	Title  string
	Url    string
}

func NewState(recentCount int) *State {
	return &State{Feeds: make(map[string]*Feed), Recent: ring.New(recentCount)}
}

func (this *State) Subscribe(url string) {
	if DEBUG {
		fmt.Fprintf(os.Stderr, "[i] SUBSCRIBE %s\n", url)
	}

	this.Lock()
	_, exists := this.Feeds[url]
	if !exists {
		this.Feeds[url] = &Feed{rss.New(10, true, this.ChannelHandler, this.ItemHandler), 0, 0}
	}
	this.Unlock()
}

func (this *State) Unsubscribe(url string) {
	if DEBUG {
		fmt.Fprintf(os.Stderr, "[i] UNSUBSCRIBE %s\n", url)
	}
	delete(this.Feeds, url)
}

func (this *State) ChannelHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	//fmt.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func (this *State) ItemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	for idx := len(newitems) - 1; idx >= 0; idx-- {
		item := newitems[idx]
		if len(item.Links) > 0 {
			this.Feeds[feed.Url].ArticleCount += 1
			urlCh <- item.Links[0].Href
			// Add to recent items.
			this.Lock()
			this.Recent.Value = RecentItem{Source: ch.Title, Title: item.Title, Url: item.Links[0].Href}
			this.Recent = this.Recent.Next()
			this.Unlock()
		}
	}
}
