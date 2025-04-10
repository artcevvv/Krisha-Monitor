package parser

import (
	"fmt"

	"github.com/gocolly/colly"
)

func TestParser() {
	c := colly.NewCollector()

	c.OnHTML("a", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit("http://go-colly.org/")
}

func test() {
	fmt.Println("QWer")
}
