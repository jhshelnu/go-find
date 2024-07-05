package main

import (
	"container/heap"
	"sort"
	"strings"
	"unicode"
)

const maxSearchResults = 10
const minimumPercentMatch = 65

type ArticleIndex struct {
	index   map[string][]string
	scraper MediumWebScraper
}

func MakeArticleIndex() ArticleIndex {
	return ArticleIndex{
		index:   make(map[string][]string),
		scraper: MakeMediumWebScraper(),
	}
}

func (ai ArticleIndex) AddArticle(url string) error {
	articleText, err := ai.scraper.getArticleText(url)
	if err != nil {
		return err
	}

	indexedNGrams := make(map[string]bool)

	for _, ngram := range GenerateNGrams(articleText) {
		if _, exists := indexedNGrams[ngram]; exists {
			continue
		}

		if urls, exists := ai.index[ngram]; exists {
			ai.index[ngram] = append(urls, url)
		} else {
			ai.index[ngram] = []string{url}
		}

		indexedNGrams[ngram] = true
	}

	return nil
}

func (ai ArticleIndex) SearchArticles(searchText string) []UrlMatch {
	ngrams := GenerateNGrams(searchText)

	ngramMatchCounts := make(map[string]int)
	for _, ngram := range ngrams {
		urls := ai.index[ngram]

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

	// getTopN gives us the top 10 urls by percentMatch **not guaranteed to be in sorted order**
	// so a final sort on the top 10 is needed before showing to the user
	results := percentMatchHeap.getTopN(maxSearchResults)
	sort.Slice(results, func(i, j int) bool {
		return results[i].PercentMatch > results[j].PercentMatch
	})

	return results
}

func GenerateNGrams(text string) []string {
	text = Normalize(text)

	ngrams := make([]string, 0)
	for i := 0; i <= len(text)-ngramSize; i++ {
		ngrams = append(ngrams, text[i:i+ngramSize])
	}

	return ngrams
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
