package main

import (
	"log"
	"fmt"
	"regexp"
	"strings"
	"net/http"
	"io/ioutil"
	"github.com/gosimple/conf"
	"github.com/noahm/rss-proxy/rss"
)

var client *http.Client
var config *conf.Config
var pathPrefix string

const FILTER_INCLUDE = 1
const FILTER_EXCLUDE = 2

type Filter struct {
	FieldName string
	Type int
	Pattern *regexp.Regexp
}

type Feed struct {
	Name string
	url string
	username string
	password string
	agent string
	filter *Filter
}

func NewFeed(name, url, username, password, agent string) *Feed {
	f := &Feed{
		Name: name,
		url: url,
		username: username,
		password: password,
		agent: agent,
		filter: nil,
	}
	echo("Registering handler for "+pathPrefix+name)
	http.Handle(pathPrefix+name, f)
	return f
}

func (f *Feed)AddFilter(t int, field string, pattern string) {
	f.filter = &Filter{
		FieldName: field,
		Type: t,
		Pattern: regexp.MustCompile(pattern),
	}
}

func (f *Feed)ServeHTTP(respWriter http.ResponseWriter, req *http.Request) {
	echo("Proxying "+f.Name)

	// request feed from remote
	feedReq, _ := http.NewRequest("GET", f.url, nil)
	if (f.username != "") {
		feedReq.SetBasicAuth(f.username, f.password)
	}
	if (f.agent != "") {
		feedReq.Header.Set("User-Agent", f.agent)
	}
	feedResp, _ := client.Do(feedReq)
	defer feedResp.Body.Close()
	// copy headers
	for field, values := range feedResp.Header {
		if (strings.ToLower(field) != "content-length") {
			for _, value := range values {
				respWriter.Header().Add(field, value)
			}
		}
	}

	respWriter.WriteHeader(feedResp.StatusCode)
	if (feedResp.StatusCode != 200 || f.filter == nil) {
		// copy feed content without parsing
		buf, _ := ioutil.ReadAll(feedResp.Body)
		respWriter.Write(buf)
	} else {
		// parse and filter feed content
		feed, _ := rss.ParseFromReader(feedResp.Body)
		selectedItems := []rss.Item{}
		for _, item := range feed.Channel.Items {
			match := f.filter.Pattern.MatchString(item.GetField(f.filter.FieldName))
			if ( match && f.filter.Type == FILTER_INCLUDE) ||
			   (!match && f.filter.Type == FILTER_EXCLUDE) {
				selectedItems = append(selectedItems, item)
			}
		}
		feed.Channel.Items = selectedItems
		buf, _ := feed.ToBytes()
		respWriter.Write(buf)
	}
}

func unknownFeed(respWriter http.ResponseWriter, req *http.Request) {
	echo("Unknown feed requested")
	respWriter.WriteHeader(404)
}

func echo(s string) {
	fmt.Println(s)
}

func main() {
	client = &http.Client{}
	config, _ := conf.ReadFile("server.conf")

	// handle 404s for unknown feeds
	path, _ := config.String("", "path-prefix")
	pathPrefix = path+"/"
	echo("Handling unknown feeds with path "+pathPrefix)
	http.HandleFunc(pathPrefix, unknownFeed)

	// read in configured feeds
	feeds := make([]*Feed, 0)
	var newFeed *Feed
	var filterType int
	var filterPattern string
	for _, section := range config.Sections() {
		if section == "default" {
			continue
		}
		feed, _ := config.String(section, "feed")
		username, _ := config.String(section, "username")
		password, _ := config.String(section, "password")
		agent, err := config.String(section, "user-agent")
		if (err != nil) {
			agent, _ = config.String("", "user-agent")
		}
		newFeed = NewFeed(
			section,
			feed,
			username,
			password,
			agent,
		)
		filterField, _ := config.String(section, "filter-field")
		if filterField != "" {
			filterPattern, _ = config.String(section, "filter-include")
			if filterPattern == "" {
				filterPattern, _ = config.String(section, "filter-exclude")
				if filterPattern == "" {
					panic("No filter-include or filter-exclude patterns include with filter-field directive in section: "+section)
				}
				filterType = FILTER_EXCLUDE
			} else {
				filterType = FILTER_INCLUDE
			}
			newFeed.AddFilter(filterType, filterField, filterPattern)
		}
		feeds = append(feeds, newFeed)
	}

	// start server
	useSsl, _ := config.Bool("", "use-ssl")
	address, _ := config.String("", "serve-address")
	if (useSsl) {
		cert, _ := config.String("", "ssl-cert")
		key, _ := config.String("", "ssl-key")
		log.Fatal(http.ListenAndServeTLS(address, cert, key, nil))
	} else {
		log.Fatal(http.ListenAndServe(address, nil))
	}
}
