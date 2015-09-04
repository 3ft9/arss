package main

import (
	"fmt"
	"os"
	"time"
)

func feedDispatcher(state *State, poolcount int, sleeptime time.Duration) {
	sleepy := true
	chkCh := make(chan string, 1000)

	for i := 0; i < poolcount; i++ {
		go feedChecker(chkCh, state)
	}

	for {
		//fmt.Fprintf(os.Stderr, "[i] LOOPING\n")
		sleepy = true
		state.Lock()
		for url, feed := range state.Feeds {
			var now = time.Now().Unix()
			if DEBUG {
				if feed.CheckDueAt < 0 {
					fmt.Fprintf(os.Stdout, "[i] CHECKING %s [%d < %d = %d]\n", url, feed.CheckDueAt, now, now-feed.CheckDueAt)
				}
			}
			if feed.CheckDueAt != -1 && feed.CheckDueAt < now {
				if DEBUG {
					fmt.Fprintf(os.Stdout, "[i] QUEUEING %s\n", url)
				}
				Stats(CONNECT_API_KEY, CONNECT_PROJECT_ID, CONNECT_COLLECTION, map[string]interface{}{
					"feed":      url,
					"action":    "queueing",
					"overdueBy": now - feed.CheckDueAt,
				})

				feed.CheckDueAt = -1
				chkCh <- url
				sleepy = false
			}
		}
		state.Unlock()

		if sleepy {
			<-time.After(time.Duration(sleeptime * time.Second))
		}
	}
}

func feedChecker(linkChan chan string, state *State) {
	for url := range linkChan {
		if DEBUG {
			fmt.Fprintf(os.Stdout, "[i] GETTING %s\n", url)
		}
		Stats(CONNECT_API_KEY, CONNECT_PROJECT_ID, CONNECT_COLLECTION, map[string]interface{}{
			"feed":   url,
			"action": "checking",
		})

		state.Lock()
		feed := state.Feeds[url]
		feed.CheckDueAt = -1
		state.Unlock()

		if DEBUG {
			fmt.Fprintf(os.Stdout, "[i] CHECKING %s\n", url)
		}

		if err := feed.Object.Fetch(url, nil); err != nil {
			if DEBUG {
				fmt.Fprintf(os.Stderr, "[e] %s: %s", url, err)
			}
			continue
		}

		if DEBUG {
			fmt.Fprintf(os.Stdout, "[i] SCHEDULING %s for %d\n", url, feed.Object.SecondsTillUpdate())
		}

		feed.CheckDueAt = time.Now().Unix() + feed.Object.SecondsTillUpdate()
	}
}
