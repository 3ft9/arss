package main

import (
	"fmt"
	rss "github.com/jteeuwen/go-pkg-rss"
	"os"
	"sync"
)

type FeedList struct {
	sync.Mutex
	Feeds map[string]*Feed
}

type Feed struct {
	Object       *rss.Feed
	CheckDueAt   int64
	ArticleCount int64
}

func NewFeed() *Feed {
	return &Feed{rss.New(10, true, channelHandler, itemHandler), 0, 0}
}

func NewFeedList() *FeedList {
	return &FeedList{Feeds: make(map[string]*Feed)}
}

func (this *FeedList) Subscribe(url string) {
	if DEBUG {
		fmt.Fprintf(os.Stderr, "[i] SUBSCRIBE %s\n", url)
	}

	this.Lock()
	_, exists := this.Feeds[url]
	if !exists {
		this.Feeds[url] = NewFeed()
	}
	this.Unlock()
}

func (this *FeedList) Unsubscribe(url string) {
	if DEBUG {
		fmt.Fprintf(os.Stderr, "[i] UNSUBSCRIBE %s\n", url)
	}
	delete(this.Feeds, url)
}
