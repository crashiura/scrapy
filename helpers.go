package scrapy

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func Text(selection *goquery.Selection) string {
	var buf bytes.Buffer

	// Slightly optimized vs calling Each: no single selection object created
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			// Keep newlines and spaces, like jQuery
			buf.WriteString(n.Data)
		}
		if n.FirstChild != nil {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	for _, n := range selection.Nodes {
		f(n)
	}

	return buf.String()
}
