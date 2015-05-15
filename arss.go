package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/3ft9/GoOse"
	zmq "github.com/pebbe/zmq4"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var DEBUG bool
var STDOUT_OUTPUT bool
var ZMQ_PUB_ADDRESS string
var HTTP_PORT int
var RECENT_ITEM_COUNT int
var STATE_FILENAME string
var STATE_SAVE_FREQUENCY int

var state *State
var urlCh chan string
var stdoutEmitterCh chan *goose.Article
var zmqEmitterCh chan *goose.Article

type TemplateData struct {
	Message string
	State   *State
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
		fmt.Fprintf(w, "    <a href=\"https://github.com/3ft9/arss\"><img style=\"position: absolute; top: 0; right: 0; border: 0;\" src=\"https://camo.githubusercontent.com/38ef81f8aca64bb9a64448d0d70f1308ef5341ab/68747470733a2f2f73332e616d617a6f6e6177732e636f6d2f6769746875622f726962626f6e732f666f726b6d655f72696768745f6461726b626c75655f3132313632312e706e67\" alt=\"Fork me on GitHub\" data-canonical-src=\"https://s3.amazonaws.com/github/ribbons/forkme_right_darkblue_121621.png\"></a>")
		fmt.Fprintf(w, "    <h1>ARSS</h1>\n")

		fmt.Fprintf(w, "    <p>%s</p>\n", template.HTMLEscapeString(data.Message))
		fmt.Fprintf(w, "    <form method=\"post\" action=\"/\">\n")
		fmt.Fprintf(w, "    	<input type=\"text\" name=\"url\" size=\"60\" />\n")
		fmt.Fprintf(w, "    	<input type=\"submit\" name=\"op\" value=\"Subscribe\" />\n")
		fmt.Fprintf(w, "    </form>\n")

		fmt.Fprintf(w, "    <h2>Feeds</h2>\n")
		fmt.Fprintf(w, "    <div id=\"feedstable\">\n")
	}

	if len(data.State.Feeds) == 0 {
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
		data.State.Lock()
		for k, v := range data.State.Feeds {
			key := ""
			if len(v.Object.Channels) > 0 {
				key += v.Object.Channels[0].Title
			}
			key += "&|&|&|&|&" + k
			keys = append(keys, key)
		}
		data.State.Unlock()
		sort.Strings(keys)

		var totalcount int64 = 0

		for _, url := range keys {
			url = strings.Split(url, "&|&|&|&|&")[1]
			totalcount += data.State.Feeds[url].ArticleCount

			fmt.Fprintf(w, "      <tr>\n")
			if len(data.State.Feeds[url].Object.Channels) > 0 {
				fmt.Fprintf(w, "        <td>%s</td>\n", template.HTMLEscapeString(data.State.Feeds[url].Object.Channels[0].Title))
			} else {
				fmt.Fprintf(w, "        <td>-</td>\n")
			}
			fmt.Fprintf(w, "        <td>%s</td>\n", template.HTMLEscapeString(url))
			fmt.Fprintf(w, "        <td align=\"right\">%d</td>\n", data.State.Feeds[url].ArticleCount)
			due := data.State.Feeds[url].CheckDueAt - now
			// if due < 5 {
			// 	fmt.Fprintf(w, "        <td align=\"right\">due now</td>\n")
			// } else {
			fmt.Fprintf(w, "        <td align=\"right\">%d secs</td>\n", due)
			// }
			fmt.Fprintf(w, "        <td><form method=\"post\" action=\"/\"><input type=\"hidden\" name=\"url\" value=\"%s\" /><input type=\"submit\" name=\"op\" value=\"Unsubscribe\" /></form></td>\n", template.HTMLEscapeString(url))
			fmt.Fprintf(w, "      </tr>\n")
		}

		fmt.Fprintf(w, "      <tr>\n")
		fmt.Fprintf(w, "        <td colspan=\"2\" style=\"border-left: none; border-bottom: none; font-weight: bold;\" align=\"right\">Total:</td>\n")
		fmt.Fprintf(w, "        <td align=\"right\">%d</td>\n", totalcount)
		fmt.Fprintf(w, "      </tr>\n")
		fmt.Fprintf(w, "    </table>\n")
	}

	fmt.Fprintf(w, "    <h2>Recent Items</h2>")
	fmt.Fprintf(w, "    <ul>")
	items := []string{}
	data.State.Lock()
	data.State.Recent.Do(func(ringitem interface{}) {
		if ringitem != nil {
			item := ringitem.(RecentItem)
			items = append(items, fmt.Sprintf("        <li><a href=\"%s\" target=\"_blank\"\">%s: %s</a></li>", template.HTMLEscapeString(item.Url), template.HTMLEscapeString(item.Source), template.HTMLEscapeString(item.Title)))
		}
	})
	data.State.Unlock()
	if len(items) == 0 {
		fmt.Fprintf(w, "        <li>None, yet.</li>")
	} else {
		for idx := len(items) - 1; idx >= 0; idx-- {
			fmt.Fprintf(w, items[idx])
		}
	}
	fmt.Fprintf(w, "    </ul>")

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
				state.Subscribe(url)
				http.Redirect(w, r, "/", 302)
			} else if op == "Unsubscribe" {
				state.Unsubscribe(url)
				http.Redirect(w, r, "/", 302)
			} else {
				msg = "Unknown operation: [" + op + "]"
			}
		}
	}

	displayTemplate(w, &TemplateData{Message: msg, State: state}, r.URL.Query().Get("ajax") == "1")
}

func articleProcessor() {
	g := goose.New()
	for url := range urlCh {
		article := g.ExtractFromUrl(url)
		if STDOUT_OUTPUT {
			stdoutEmitterCh <- article
		}
		if len(ZMQ_PUB_ADDRESS) > 0 {
			zmqEmitterCh <- article
		}
	}
}

func stdoutEmitter() {
	e := json.NewEncoder(os.Stdout)
	for article := range stdoutEmitterCh {
		e.Encode(article)
	}
}

func zmqEmitter() {
	publisher, _ := zmq.NewSocket(zmq.PUB)
	defer publisher.Close()
	publisher.Bind(ZMQ_PUB_ADDRESS)
	for article := range zmqEmitterCh {
		output, err := json.Marshal(article)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[e] %s\n", err)
		} else {
			publisher.Send(string(output), 0)
		}
	}
}

func main() {
	flag.BoolVar(&DEBUG, "debug", false, "Enable debugging output")
	flag.BoolVar(&STDOUT_OUTPUT, "stdout", false, "Output to stdout")
	flag.StringVar(&ZMQ_PUB_ADDRESS, "zmq_pub_address", "", "ZMQ publish address, default is none (disabled)")
	flag.IntVar(&HTTP_PORT, "port", 8080, "HTTP server port")
	flag.IntVar(&RECENT_ITEM_COUNT, "recent", 20, "Number of recent items to display")
	flag.StringVar(&STATE_FILENAME, "state_filename", "arss.state", "Filename in which to store the state")
	flag.IntVar(&STATE_SAVE_FREQUENCY, "state_frequency", 60, "Frequency with which the state is saved (seconds)")
	flag.Parse()

	// Single stdoutEmitter to control access.
	if STDOUT_OUTPUT {
		stdoutEmitterCh = make(chan *goose.Article, 1000)
		go stdoutEmitter()
	}

	// Single ZMQ to control access.
	if len(ZMQ_PUB_ADDRESS) > 0 {
		zmqEmitterCh = make(chan *goose.Article, 1000)
		go zmqEmitter()
	}

	urlCh = make(chan string, 1000)
	for i := 0; i < 4; i++ {
		go articleProcessor()
	}

	state = NewState(RECENT_ITEM_COUNT, STATE_FILENAME, STATE_SAVE_FREQUENCY)
	// State.Subscribe("http://feeds.bbci.co.uk/news/rss.xml")
	// State.Subscribe("http://feeds.mashable.com/Mashable")
	// State.Subscribe("http://feeds2.feedburner.com/techradar/allnews")
	// State.Subscribe("http://feeds.feedburner.com/TechCrunch/")
	// State.Subscribe("http://www.cnet.com/rss/news/")
	// State.Subscribe("http://www.computerweekly.com/rss/All-Computer-Weekly-content.xml")
	// State.Subscribe("http://www.zdnet.com/news/rss.xml")
	// State.Subscribe("http://feeds.wired.com/wired/index")

	go feedDispatcher(state, 4, 10)

	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":"+strconv.Itoa(HTTP_PORT), nil)
}
