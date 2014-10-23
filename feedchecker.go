package main

import (
	"fmt"
	"os"
	"time"
)

func feedDispatcher(feeds *FeedList, poolcount int, sleeptime time.Duration) {
	sleepy := true
	chkCh := make(chan string, 1000)

	for i := 0; i < poolcount; i++ {
		go feedChecker(chkCh, feeds)
	}

	for {
		//fmt.Fprintf(os.Stderr, "[i] LOOPING\n")
		sleepy = true
		feeds.Lock()
		for url, feed := range feeds.Feeds {
			var now = time.Now().Unix()
			if feed.CheckDueAt != -1 && feed.CheckDueAt < now {
				feed.CheckDueAt = -1
				chkCh <- url
				sleepy = false
			}
		}
		feeds.Unlock()

		if sleepy {
			<-time.After(time.Duration(sleeptime * time.Second))
		}
	}
}

func feedChecker(linkChan chan string, feeds *FeedList) {
	for url := range linkChan {
		if DEBUG {
			fmt.Fprintf(os.Stderr, "[i] GETTING %s\n", url)
		}

		feeds.Lock()
		feed := feeds.Feeds[url]
		feed.CheckDueAt = -1
		feeds.Unlock()

		if DEBUG {
			fmt.Fprintf(os.Stderr, "[i] CHECKING %s\n", url)
		}

		if err := feed.Object.Fetch(url, nil); err != nil {
			if DEBUG {
				fmt.Fprintf(os.Stderr, "[e] %s: %s", url, err)
			}
			continue
		}

		if DEBUG {
			fmt.Fprintf(os.Stderr, "[i] SCHEDULING %s for %d\n", url, feed.Object.SecondsTillUpdate())
		}

		feed.CheckDueAt = time.Now().Unix() + feed.Object.SecondsTillUpdate()
	}
}
