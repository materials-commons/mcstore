package main

import (
	"fmt"
	"io/ioutil"

	"golang.org/x/net/html"
	"bytes"
)

func main() {
	b, err := ioutil.ReadFile("/Users/gtarcea/Dropbox/transfers/g/x.map")
	if err != nil {
		fmt.Println("Unable to read file:", err)
	}
	doc, err := html.Parse(bytes.NewBuffer(b))
	if err != nil {
		fmt.Println("Failed to parse html:", err)
	}
	buildMap(doc)
}

func buildMap(node *html.Node) {
	if node.Type == html.ElementNode && node.Data == "area" {
		for _, attr := range node.Attr {
			fmt.Println(attr.Key, attr.Val)

		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		buildMap(c)
	}
}
