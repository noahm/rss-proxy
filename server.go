package main

import (
	"log"
	"fmt"
	"net/http"
	"io/ioutil"
	"github.com/gosimple/conf"
	"github.com/noahm/rss-proxy/rss"
)

var client *http.Client
var config *conf.Config
var pathPrefix string

// type Filter struct {
// 	Field string
// 	Pattern string
// }

type Feed struct {
	Name string
	url string
	username string
	password string
	// filter Filter
}

func NewFeed(name, url, username, password string) *Feed {
	f := &Feed{
		Name: name,
		url: url,
		username: username,
		password: password,
	}
	echo("Registering handler for "+pathPrefix+name)
	http.Handle(pathPrefix+name, f)
	return f
}

func (f *Feed)ServeHTTP(respWriter http.ResponseWriter, req *http.Request) {
	echo("Proxying "+f.Name)

	// request feed from remote
	feedReq, _ := http.NewRequest("GET", f.url, nil)
	feedReq.SetBasicAuth(f.username, f.password)
	feedResp, _ := client.Do(feedReq)
	defer feedResp.Body.Close()
	// copy headers
	for field, values := range feedResp.Header {
		for _, value := range values {
			respWriter.Header().Add(field, value)
		}
	}

	// copy feed content
	respWriter.WriteHeader(feedResp.StatusCode)
	if (feedResp.StatusCode != 200) {
		buf, _ := ioutil.ReadAll(feedResp.Body)
		respWriter.Write(buf)
	} else {
		feed, _ := rss.ParseFromReader(feedResp.Body)
		fmt.Printf("First item: %v\r\n", feed.Channel.Items[0])
		// selectedItems := []rss.Item{}
		// for _, item := range feed.Channel.Items {
		// 	if item.Guid == "http://v.giantbomb.com/2015/08/14/mc_mgs04_ep15_08132015_34lqov67_4000.mp4" {
		// 		selectedItems = append(selectedItems, item)
		// 	}
		// }
		// feed.Channel.Items = selectedItems
		buf, _ := feed.ToBytes()
		respWriter.Write(buf)
	}
}

func unknownFeed(respWriter http.ResponseWriter, req *http.Request) {
	echo("Unknown feed requested")
	respWriter.WriteHeader(404)
}

func echo(s string) {
	fmt.Println(s+"\r")
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
	for _, section := range config.Sections() {
		if section == "default" {
			continue
		}
		feed, _ := config.String(section, "feed")
		username, _ := config.String(section, "username")
		password, _ := config.String(section, "password")
		feeds = append(feeds, NewFeed(
			section,
			feed,
			username,
			password,
		))
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
