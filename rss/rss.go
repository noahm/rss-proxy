// based on github.com/krautchan/gbt/module/api/rss
//
// "THE PIZZA-WARE LICENSE" (derived from "THE BEER-WARE LICENCE"):
// <whoami@dev-urandom.eu> wrote these files. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you think
// this stuff is worth it, you can buy me a pizza in return.

/*
Package to parse and create RSS-Feeds
*/
package rss

import (
    "encoding/xml"
    "io"
    "reflect"
)

type Rss struct {
    XMLName  xml.Name `xml:"rss"`
    Version  string   `xml:"version,attr"`
    Channel  Channel  `xml:"channel"`
}

type Channel struct {
    Title         string `xml:"title"`
    Description   string `xml:"description"`
    Link          string `xml:"link"`
    Image         Image  `xml:"image"`
    Items         []Item `xml:"item"`
}

type Image struct {
    Title string `xml:"title"`
    Url   string `xml:"url"`
    Link  string `xml:"link"`
    Width  string `xml:"width"`
    Height  string `xml:"height"`
}

type Item struct {
    Title       string `xml:"title"`
    Link        string `xml:"link"`
    Description string `xml:"description"`
    PupDate     string `xml:"pubDate"`
    Guid        string `xml:"guid"`
    Duration    string `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd duration"`
    Image       string `xml:"http://www.itunes.com/dtds/podcast-1.0.dtd image"`
    Enclosures  []Enc  `xml:"enclosure"`
}

func (item *Item) GetField(fieldName string) string {
    r := reflect.ValueOf(item)
    f := reflect.Indirect(r).FieldByName(fieldName)
    return string(f.String())
}

type Enc struct {
    Url    string `xml:"url,attr"`
    Type   string `xml:"type,attr"`
    Length string `xml:"length,attr"`
}

func ParseFromReader(reader io.Reader) (*Rss, error) {
    var rss Rss
    dec := xml.NewDecoder(reader)
    err := dec.Decode(&rss)
    if err != nil {
        return nil, err
    }
    return &rss, nil
}

// Writes the data in RSS 2.0 format to a given ResponseWriter object
func (rss *Rss) ToBytes() ([]byte, error) {
    bytes, err := xml.MarshalIndent(rss, "", "  ")
    if err != nil {
        return nil, err
    }
    bytes = append([]byte(xml.Header), bytes...)
    return bytes, nil
}
