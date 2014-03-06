package main

import (
	"log"
	"net/http"

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
	// Fetch the URL.
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Parse the HTML to find the <title> contents.
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return err
	}
	title := doc.Find("title").Text()
	if title == "" {
		title = "(untitled)"
	}

	// Add the new link.
	link := &link{
		Title: title,
		URL:   url,
	}
	err = addAndBroadcastLink(link)
	if err != nil {
		log.Printf("Error adding and broadcasting link %v: %s", link, err)
	}

	return nil
}
