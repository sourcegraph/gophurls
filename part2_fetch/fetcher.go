package main

import (
	"log"

	"github.com/PuerkitoBio/goquery"
)

func startFetcher() {
	for url := range urlsToFetch {
		err := fetch(url)
		if err != nil {
			log.Printf("Error fetching %q: %s", url, err)
		}
	}
}

func fetch(url string) error {
	// Fetch and parse the HTML to find the <title> contents.
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return err
	}
	title := doc.Find("title").Text()
	if title == "" {
		title = "(untitled)"
	}

	// Lock links for writing and add the new link.
	linksMutex.Lock()
	defer linksMutex.Unlock()
	links = append(links, &link{
		Title: title,
		URL:   url,
	})

	return nil
}
