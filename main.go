package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

const ngramSize = 5
const genericError = "Expected \"add [medium article link]\", \"search [search term]\", or \"quit\""

// TODO:
//   - web scraping seems to miss some bits of the article, consider relaxing selectors
//   - handle junk search queries, e.g. "search ;;;;;;;;" where all of the query chars get normalized out. (validate on input that there's at least 5 GOOD chars in there)
//   - explore a recursive indexing approach where links to other medium articles are followed, to easily index a subtree of medium's articles
func main() {
	articleIndex := MakeArticleIndex()
	scanner := bufio.NewScanner(os.Stdin)

	// todo: temporarily pre-index a couple articles to help with testing
	_ = articleIndex.AddArticle("https://prathamrathour2018.medium.com/golang-vs-java-an-in-depth-comparison-07a2569ca2ee")
	_ = articleIndex.AddArticle("https://medium.com/@ahmed.nums345/a-comprehensive-guide-to-next-js-5f3b03b49def")

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

			err := articleIndex.AddArticle(arg)
			if err != nil && err.Error() != "URL already visited" {
				fmt.Println(genericError)
			}
		case "search":
			if len(arg) < ngramSize {
				fmt.Printf("expected at least %d search characters!\n", ngramSize)
				continue
			}

			DisplaySearchResults(articleIndex.SearchArticles(arg))
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

func DisplaySearchResults(results []UrlMatch) {
	if len(results) == 0 {
		fmt.Print("Your search returned 0 results\n\n")
		return
	}

	for idx, result := range results {
		fmt.Printf("%d: %s (%.2f%% match)\n", idx+1, result.Url, result.PercentMatch)
	}
	fmt.Println()
}
