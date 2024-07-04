package main

import (
	"bufio"
	"container/heap"
	"errors"
	"fmt"
	"github.com/gocolly/colly"
	"os"
	"regexp"
	"strings"
	"unicode"
)

const ngramSize = 5
const maxSearchResults = 10
const minimumPercentMatch = 65
const genericError = "Expected \"add [medium article link]\", \"search [search term]\", or \"quit\""

var mediumRegexp = regexp.MustCompile("https://(.*\\.)?medium.com(/.*)?")
var scanner = bufio.NewScanner(os.Stdin)

// TODO:
//   - web scraping seems to miss some bits of the article, consider relaxing selectors
//   - extract articleIndex into a struct which can use an additional map[string]bool to keep track of already indexed urls (prevents reindexing). can expose methods for IndexArticle and SearchArticles
//   - handle junk search queries, e.g. "search ;;;;;;;;" where all of the query chars get normalized out. (validate on input that there's at least 5 GOOD chars in there)
//   - explore a recursive indexing approach where links to other medium articles are followed, to easily index a subtree of medium's articles
func main() {
	// maps a single ngram to a list of articles which contain that ngram
	articleIndex := make(map[string][]string)

	// todo: temporarily pre-index a couple articles to help with testing
	_ = IndexArticle("https://prathamrathour2018.medium.com/golang-vs-java-an-in-depth-comparison-07a2569ca2ee", &articleIndex)
	_ = IndexArticle("https://medium.com/@ahmed.nums345/a-comprehensive-guide-to-next-js-5f3b03b49def", &articleIndex)

	for {
		fmt.Print("> ")

		command, arg, err := ParseInput(scanner)
		if err != nil {
			fmt.Println(genericError)
			continue
		}

		switch command {
		case "add":
			if len(arg) == 0 {
				fmt.Println(genericError)
				continue
			}

			err := IndexArticle(arg, &articleIndex)
			if err != nil {
				fmt.Println(genericError)
			}
		case "search":
			if len(arg) < ngramSize {
				fmt.Printf("expected at least %d search characters!\n", ngramSize)
				continue
			}

			SearchArticles(arg, articleIndex)
		case "quit":
			return
		default:
			fmt.Println(genericError)
		}
	}
}

// ParseInput reads in one line from stdin and parses it into a command (the first word) and an argument (everything after a space)
func ParseInput(scanner *bufio.Scanner) (string, string, error) {
	scanner.Scan()
	line := strings.SplitN(scanner.Text(), " ", 2)
	var command, arg string

	if len(line) == 0 {
		return "", "", errors.New("received no input")
	}

	command = strings.ToLower(line[0])
	if len(command) == 0 {
		return "", "", errors.New("received leading space")
	}

	if len(line) == 2 {
		arg = strings.ToLower(strings.TrimSpace(line[1]))
	}

	return command, arg, nil
}

func IndexArticle(url string, articleIndex *map[string][]string) error {
	articleText := ""
	scraper := colly.NewCollector(colly.URLFilters(mediumRegexp))
	scraper.OnHTML("article p.pw-post-body-paragraph", func(e *colly.HTMLElement) {
		articleText += " " + e.Text
	})

	err := scraper.Visit(url)
	if err != nil {
		return err
	}

	indexedNgrams := make(map[string]bool)

	for _, ngram := range GenerateNGrams(articleText) {
		if _, exists := indexedNgrams[ngram]; exists {
			continue
		}

		if urls, exists := (*articleIndex)[ngram]; exists {
			(*articleIndex)[ngram] = append(urls, url)
		} else {
			(*articleIndex)[ngram] = []string{url}
		}

		indexedNgrams[ngram] = true
	}

	return nil
}

func SearchArticles(searchText string, articleIndex map[string][]string) {
	ngrams := GenerateNGrams(searchText)

	ngramMatchCounts := make(map[string]int)
	for _, ngram := range ngrams {
		urls := articleIndex[ngram]

		for _, url := range urls {
			if _, exists := ngramMatchCounts[url]; exists {
				ngramMatchCounts[url]++
			} else {
				ngramMatchCounts[url] = 1
			}
		}
	}

	percentMatchHeap := make(UrlMatchHeap, 0, len(ngramMatchCounts))
	heap.Init(&percentMatchHeap)

	for url, ngramsMatched := range ngramMatchCounts {
		percentMatch := float32(ngramsMatched) * 100 / float32(len(ngrams))
		if percentMatch > minimumPercentMatch {
			heap.Push(&percentMatchHeap, UrlMatch{url, percentMatch})
		}
	}

	DisplaySearchResults(percentMatchHeap.getTopN(maxSearchResults))
}

func DisplaySearchResults(results []UrlMatch) {
	for idx, result := range results {
		fmt.Printf("%d: %s (%.2f%% match)\n", idx+1, result.Url, result.PercentMatch)
	}
	fmt.Println()
}

func Normalize(text string) string {
	var normalizedText strings.Builder
	for _, char := range text {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			normalizedText.WriteRune(unicode.ToLower(char))
		}
	}
	return normalizedText.String()
}

func GenerateNGrams(text string) []string {
	ngrams := make([]string, 0)

	text = Normalize(text)

	for i := 0; i <= len(text)-ngramSize; i++ {
		ngrams = append(ngrams, text[i:i+ngramSize])
	}

	return ngrams
}
