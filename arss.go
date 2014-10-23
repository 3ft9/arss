package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/3ft9/GoOse"
	rss "github.com/jteeuwen/go-pkg-rss"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

var DEBUG bool
var STDOUT_OUTPUT bool

var feedlist *FeedList
var urlCh chan string

type TemplateData struct {
	Message string
	Feeds   *FeedList
}

func displayTemplate(w http.ResponseWriter, data *TemplateData, tableonly bool) {
	w.Header().Add("Content-Type", "text/html")

	if !tableonly {
		fmt.Fprintf(w, "<!DOCTYPE html>\n")
		fmt.Fprintf(w, "<html>\n")
		fmt.Fprintf(w, "  <head>\n")
		fmt.Fprintf(w, "    <title>ARSS</title>\n")
		fmt.Fprintf(w, "    <style type=\"text/css\">\n")
		fmt.Fprintf(w, "    table {\n")
		fmt.Fprintf(w, "    	font-family: verdana,arial,sans-serif;\n")
		fmt.Fprintf(w, "    	font-size:11px;\n")
		fmt.Fprintf(w, "    	color:#333333;\n")
		fmt.Fprintf(w, "    	border-width: 1px;\n")
		fmt.Fprintf(w, "    	border-color: #666666;\n")
		fmt.Fprintf(w, "    	border-collapse: collapse;\n")
		fmt.Fprintf(w, "    }\n")
		fmt.Fprintf(w, "    table th {\n")
		fmt.Fprintf(w, "    	border-width: 1px;\n")
		fmt.Fprintf(w, "    	padding: 8px;\n")
		fmt.Fprintf(w, "    	border-style: solid;\n")
		fmt.Fprintf(w, "    	border-color: #666666;\n")
		fmt.Fprintf(w, "    	background-color: #dedede;\n")
		fmt.Fprintf(w, "    }\n")
		fmt.Fprintf(w, "    table td {\n")
		fmt.Fprintf(w, "    	border-width: 1px;\n")
		fmt.Fprintf(w, "    	padding: 8px;\n")
		fmt.Fprintf(w, "    	border-style: solid;\n")
		fmt.Fprintf(w, "    	border-color: #666666;\n")
		fmt.Fprintf(w, "    	background-color: #ffffff;\n")
		fmt.Fprintf(w, "    }\n")
		fmt.Fprintf(w, "    </style>\n")
		fmt.Fprintf(w, "  </head>\n")
		fmt.Fprintf(w, "  <body>\n")
		fmt.Fprintf(w, "    <h1>ARSS</h1>\n")

		fmt.Fprintf(w, "    <p>%s</p>\n", template.HTMLEscapeString(data.Message))
		fmt.Fprintf(w, "    <form method=\"post\" action=\"/\">\n")
		fmt.Fprintf(w, "    	<input type=\"text\" name=\"url\" size=\"60\" />\n")
		fmt.Fprintf(w, "    	<input type=\"submit\" name=\"op\" value=\"Subscribe\" />\n")
		fmt.Fprintf(w, "    </form>\n")

		fmt.Fprintf(w, "    <h2>Feeds</h2>\n")
		fmt.Fprintf(w, "    <div id=\"feedstable\">\n")
	}

	if len(data.Feeds.Feeds) == 0 {
		fmt.Fprintf(w, "    <p>None, yet.</p>\n")
	} else {
		fmt.Fprintf(w, "    <table>\n")
		fmt.Fprintf(w, "      <tr>\n")
		fmt.Fprintf(w, "        <th>Name</th>\n")
		fmt.Fprintf(w, "        <th>URL</th>\n")
		fmt.Fprintf(w, "        <th>Count</th>\n")
		fmt.Fprintf(w, "        <th align=\"right\">Check due</th>\n")
		fmt.Fprintf(w, "      </tr>\n")

		now := time.Now().Unix()
		var keys []string
		for k, v := range data.Feeds.Feeds {
			key := ""
			if len(v.Object.Channels) > 0 {
				key += v.Object.Channels[0].Title
			}
			key += "&|&|&|&|&" + k
			keys = append(keys, key)
		}
		sort.Strings(keys)

		var totalcount int64 = 0

		for _, url := range keys {
			url = strings.Split(url, "&|&|&|&|&")[1]
			totalcount += data.Feeds.Feeds[url].ArticleCount

			fmt.Fprintf(w, "      <tr>\n")
			if len(data.Feeds.Feeds[url].Object.Channels) > 0 {
				fmt.Fprintf(w, "        <td>%s</td>\n", template.HTMLEscapeString(data.Feeds.Feeds[url].Object.Channels[0].Title))
			} else {
				fmt.Fprintf(w, "        <td>-</td>\n")
			}
			fmt.Fprintf(w, "        <td>%s</td>\n", template.HTMLEscapeString(url))
			fmt.Fprintf(w, "        <td align=\"right\">%d</td>\n", data.Feeds.Feeds[url].ArticleCount)
			due := data.Feeds.Feeds[url].CheckDueAt - now
			if due < 5 {
				fmt.Fprintf(w, "        <td align=\"right\">due now</td>\n")
			} else {
				fmt.Fprintf(w, "        <td align=\"right\">%d secs</td>\n", data.Feeds.Feeds[url].CheckDueAt-now)
			}
			fmt.Fprintf(w, "        <td><form method=\"post\" action=\"/\"><input type=\"hidden\" name=\"url\" value=\"%s\" /><input type=\"submit\" name=\"op\" value=\"Unsubscribe\" /></form></td>\n", template.HTMLEscapeString(url))
			fmt.Fprintf(w, "      </tr>\n")
		}

		fmt.Fprintf(w, "      <tr>\n")
		fmt.Fprintf(w, "        <td colspan=\"2\" style=\"border-left: none; border-bottom: none; font-weight: bold;\" align=\"right\">Total:</td>\n")
		fmt.Fprintf(w, "        <td align=\"right\">%d</td>\n", totalcount)
		fmt.Fprintf(w, "      </tr>\n")
		fmt.Fprintf(w, "    </table>\n")
	}

	if !tableonly {
		fmt.Fprintf(w, "    </div>\n")
		fmt.Fprintf(w, "    <script src=\"https://code.jquery.com/jquery-1.11.1.min.js\"></script>\n")
		fmt.Fprintf(w, "    <script type=\"text/javascript\">\n")
		fmt.Fprintf(w, "      (function updateTable() {\n")
		fmt.Fprintf(w, "        $.ajax({\n")
		fmt.Fprintf(w, "          url: '/?ajax=1',\n")
		fmt.Fprintf(w, "          success: function(data) {\n")
		fmt.Fprintf(w, "            $('#feedstable').html(data);\n")
		fmt.Fprintf(w, "          },\n")
		fmt.Fprintf(w, "          complete: function() {\n")
		fmt.Fprintf(w, "            setTimeout(updateTable, 5000);\n")
		fmt.Fprintf(w, "          }\n")
		fmt.Fprintf(w, "        });\n")
		fmt.Fprintf(w, "      })();\n")
		fmt.Fprintf(w, "    </script>\n")
		fmt.Fprintf(w, "  </body>\n")
		fmt.Fprintf(w, "</html>\n")
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	msg := "Welcome to the Automated RSS service."
	if r.Method == "POST" {
		url := r.PostFormValue("url")
		op := r.PostFormValue("op")
		if len(url) == 0 {
			msg = "Please specify a URL!"
		} else {
			if op == "Subscribe" {
				feedlist.Subscribe(url)
				http.Redirect(w, r, "/", 302)
			} else if op == "Unsubscribe" {
				feedlist.Unsubscribe(url)
				http.Redirect(w, r, "/", 302)
			} else {
				msg = "Unknown operation: [" + op + "]"
			}
		}
	}

	feedlist.Lock()
	displayTemplate(w, &TemplateData{Message: msg, Feeds: feedlist}, r.URL.Query().Get("ajax") == "1")
	feedlist.Unlock()
}

func channelHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	//fmt.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	for _, item := range newitems {
		if len(item.Links) > 0 {
			feedlist.Feeds[feed.Url].ArticleCount += 1
			urlCh <- item.Links[0].Href
		}
	}
}

func articleProcessor() {
	e := json.NewEncoder(os.Stdout)
	g := goose.New()
	for url := range urlCh {
		if STDOUT_OUTPUT {
			article := g.ExtractFromUrl(url)
			e.Encode(article)
		}
	}
}

func main() {
	flag.BoolVar(&DEBUG, "debug", false, "Enable debugging output")
	flag.BoolVar(&STDOUT_OUTPUT, "stdout", false, "Output to stdout")
	flag.Parse()

	urlCh = make(chan string, 1000)
	for i := 0; i < 2; i++ {
		go articleProcessor()
	}

	feedlist = NewFeedList()
	// feedlist.Subscribe("http://feeds.bbci.co.uk/news/rss.xml")
	// feedlist.Subscribe("http://feeds.mashable.com/Mashable")
	// feedlist.Subscribe("http://feeds2.feedburner.com/techradar/allnews")
	// feedlist.Subscribe("http://feeds.feedburner.com/TechCrunch/")
	// feedlist.Subscribe("http://www.cnet.com/rss/news/")
	// feedlist.Subscribe("http://www.computerweekly.com/rss/All-Computer-Weekly-content.xml")
	// feedlist.Subscribe("http://www.zdnet.com/news/rss.xml")
	// feedlist.Subscribe("http://feeds.wired.com/wired/index")

	go feedDispatcher(feedlist, 4, 10)

	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8080", nil)
}
