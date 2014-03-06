package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"text/template"
)

var httpAddr = flag.String("http", ":7000", "HTTP service address")
var verbose = flag.Bool("v", false, "show verbose output")

type link struct {
	URL string

	// Title is the page's title.
	Title string
}

var (
	// links stores the fetched URLs and their titles.
	links []*link

	linksMutex sync.RWMutex

	// newLinks contains newly added links (with fetched titles) that should be
	// broadcasted and added to the list of links.
	newLinks = make(chan *link)

	// urls contains newly added URLs enqueued to fetch.
	urlsToFetch = make(chan string)
)

func init() {
	// Set up the HTTP handler in init (not main) so we can test it. (This main
	// doesn't run when testing.)
	http.HandleFunc("/peers", addPeers)
	http.HandleFunc("/links", addLink)
	http.HandleFunc("/", home)

	go startFetcher()
}

func main() {
	flag.Parse()
	if *verbose {
		log.SetPrefix(*httpAddr + ": ")
	}
	if err := http.ListenAndServe(*httpAddr, nil); err != nil {
		log.Fatal(err)
	}
}

var homeTmpl = template.Must(template.New("Home").Parse(`
<h1>GophURLs <img width=28 height=38 src="http://golang.org/doc/gopher/frontpage.png"></h1>
<p>Submit a link: <tt>curl -X POST -d '{"URL":"http://example.com","Title":"optional title"}' http://localhost:7000/links</tt></p>
<p>Newly added links without titles appear only after they've been fetched.</p>
<h2>Links ({{len .}})</h2>
<ol>
{{/* Iterate over the data we passed to the template (links). */}}
{{range .}}
  <li><a href="{{.URL}}">{{.Title}}</a></li>
{{end}}
</ol>
`))

func home(w http.ResponseWriter, r *http.Request) {
	// Lock links for reading.
	linksMutex.RLock()
	defer linksMutex.RUnlock()

	// Render the template.
	homeTmpl.Execute(w, links)
}

func addLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "bad method", http.StatusBadRequest)
		return
	}

	var link *link
	err := json.NewDecoder(r.Body).Decode(&link)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad JSON: %s", err), http.StatusBadRequest)
		return
	}

	// Validate the URL.
	if link.URL == "" {
		http.Error(w, "no url", http.StatusBadRequest)
		return
	}
	url, err := url.Parse(link.URL)
	if err != nil {
		http.Error(w, "bad url", http.StatusBadRequest)
		return
	}
	if !url.IsAbs() {
		http.Error(w, "url must be absolute", http.StatusBadRequest)
		return
	}

	// Check if we already have this link. Use an anonymous function so we can
	// have `defer` execute immediately after this block.
	present := func() bool {
		linksMutex.RLock()
		defer linksMutex.RUnlock()
		for _, existingLink := range links {
			if link.URL == existingLink.URL {
				return true // nothing to do
			}
		}
		return false
	}()
	if present {
		return
	}

	if link.Title == "" {
		// If no title is specified, enqueue it to fetch the title.
		if *verbose {
			log.Printf("New link (needs fetch): %q.", link.URL)
		}
		urlsToFetch <- link.URL
	} else {
		if *verbose {
			log.Printf("New link (no fetch needed): %q.", link.URL)
		}
		err := addAndBroadcastLink(link)
		if err != nil {
			log.Printf("Error adding and broadcasting link %v: %s", link, err)
		}
	}

	if *verbose {
		log.Printf("Added new link: %s.", link.URL)
	}
}

func addAndBroadcastLink(link *link) error {
	func() {
		linksMutex.Lock()
		defer linksMutex.Unlock()
		links = append(links, link)
	}()

	return broadcastLink(link)
}
