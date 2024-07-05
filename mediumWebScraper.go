package main

import (
	"github.com/gocolly/colly"
	"regexp"
	"strings"
)

var mediumRegexp = regexp.MustCompile("https://(.*\\.)?medium.com(/.*)?")
var mediumArticleTextSelector = "article p.pw-post-body-paragraph"

type MediumWebScraper struct {
	delegate *colly.Collector
	buffer   *string // what the delegate web scraper (colly) will write to (note: not thread safe!)
}

func MakeMediumWebScraper() MediumWebScraper {
	delegate := colly.NewCollector(colly.URLFilters(mediumRegexp))
	buffer := new(string)
	mediumWebScraper := MediumWebScraper{delegate: delegate, buffer: buffer}
	delegate.OnHTML(mediumArticleTextSelector, func(e *colly.HTMLElement) {
		*mediumWebScraper.buffer += " " + e.Text
	})
	return mediumWebScraper
}

func (scraper MediumWebScraper) getArticleText(url string) (string, error) {
	*scraper.buffer = ""

	err := scraper.delegate.Visit(url)
	if err != nil {
		return "", err
	}

	return strings.Clone(*scraper.buffer), nil
}
