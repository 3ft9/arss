package main

import (
	"bufio"
	ring "container/ring"
	"fmt"
	rss "github.com/jteeuwen/go-pkg-rss"
	"io"
	"os"
	"sync"
	"time"
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
	Guids        *ring.Ring
}

type RecentItem struct {
	Source string
	Title  string
	Url    string
}

func NewState(recentCount int, filename string, frequency int) *State {
	retval := &State{Feeds: make(map[string]*Feed), Recent: ring.New(recentCount)}
	retval.Load(filename)
	go retval.PeriodicSave(filename, frequency)
	return retval
}

func (this *State) Subscribe(url string) {
	if DEBUG {
		fmt.Fprintf(os.Stderr, "[i] SUBSCRIBE %s\n", url)
	}

	this.Lock()
	this.SubscribeNoLock(url)
	this.Unlock()
}

func (this *State) SubscribeNoLock(url string) {
	_, exists := this.Feeds[url]
	if !exists {
		this.Feeds[url] = &Feed{Object: rss.New(10, true, this.ChannelHandler, this.ItemHandler), CheckDueAt: 0, ArticleCount: 0, Guids: ring.New(1000)}
	}
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

		// Have we seen it before?
		isNew := true
		guid := item.Key()
		this.Feeds[feed.Url].Guids.Do(func(ringitem interface{}) {
			if isNew && ringitem != nil {
				if ringitem.(string) == guid {
					isNew = false
				}
			}
		})

		// If we haven't seen it before, process it.
		if isNew {
			this.Feeds[feed.Url].Guids.Value = guid
			this.Feeds[feed.Url].Guids = this.Feeds[feed.Url].Guids.Next()
			if len(item.Links) > 0 {
				Stats(CONNECT_API_KEY, CONNECT_PROJECT_ID, CONNECT_COLLECTION, map[string]interface{}{
					"feedUrl":   feed.Url,
					"feedTitle": ch.Title,
					"action":    "newItem",
					"article": map[string]interface{}{
						"url":   item.Links[0].Href,
						"title": item.Title,
					},
				})

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
}

func (this *State) Load(filename string) {
	fh, err := os.Open(filename)
	defer fh.Close()
	if err == nil {
		this.Lock()
		reader := bufio.NewReader(fh)
		line, err := reader.ReadString('\n')
		for {
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "[E] %s\n", err)
				}
				this.Unlock()
				return
			} else {
				url := line[0 : len(line)-1]
				this.SubscribeNoLock(url)
				for {
					line, err = reader.ReadString('\n')
					if err != nil {
						if err != io.EOF {
							fmt.Fprintf(os.Stderr, "[E] %s\n", err)
						}
						this.Unlock()
						return
					} else {
						if line[0] == '>' {
							this.Feeds[url].Guids.Value = line[1 : len(line)-1]
							this.Feeds[url].Guids = this.Feeds[url].Guids.Next()
						} else {
							break
						}
					}
				}
			}
		}
		this.Unlock()
	}
}

func (this *State) PeriodicSave(filename string, frequency int) {
	for {
		this.Save(filename)
		<-time.After(time.Duration(frequency * 1e9))
	}
}

func (this *State) Save(filename string) {
	fh, eopen := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	defer fh.Close()
	if eopen != nil {
		fmt.Fprintf(os.Stderr, "[E] Error opening state file for writing: %s\n", eopen)
	} else {
		this.Lock()
		for url, _ := range state.Feeds {
			fh.WriteString(url + "\n")
			this.Feeds[url].Guids.Do(func(ringitem interface{}) {
				if ringitem != nil {
					fh.WriteString(">" + ringitem.(string) + "\n")
				}
			})
		}
		this.Unlock()
	}
}
