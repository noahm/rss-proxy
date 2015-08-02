package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"github.com/gosimple/conf"
)

var client *http.Client

type Feed struct {
	Name string
	url string
	username string
	password string
}

func NewFeed(name, url, username, password string) *Feed {
	f := &Feed{
		Name: name,
		url: url,
		username: username,
		password: password,
	}
	return f
}

func (f *Feed)fetch() string {
	req, _ := http.NewRequest("GET", f.url, nil)
	req.SetBasicAuth(f.username, f.password)
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	buf, _ := ioutil.ReadAll(resp.Body)
	return string(buf)
}

func main() {
	client = &http.Client{}
	feeds := make([]*Feed, 0)
	c, _ := conf.ReadFile("server.conf")
	for _, section := range c.Sections() {
		if section == "default" {
			continue
		}
		feed, _ := c.String(section, "feed")
		username, _ := c.String(section, "username")
		password, _ := c.String(section, "password")
		feeds = append(feeds, NewFeed(
			section,
			feed,
			username,
			password,
		))
	}
	fmt.Println(feeds[0].fetch())
}
