package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
)

var (
	// peers stores the set of known, reachable network peers (in "host:port"
	// format).
	peers = make(map[string]struct{})

	peersMutex sync.RWMutex
)

// addPeers is an HTTP handler that takes JSON PUT data containing an array of
// peers (in "host:port" format) and adds them to the set of peers.
func addPeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "bad method", http.StatusBadRequest)
		return
	}

	if r.Body == nil {
		http.Error(w, "no body", http.StatusBadRequest)
		return
	}

	var newPeers []string
	err := json.NewDecoder(r.Body).Decode(&newPeers)
	if err != nil {
		log.Printf("Error decoding peers JSON: %s", err)
		http.Error(w, "decode error", http.StatusBadRequest)
		return
	}

	for _, peer := range newPeers {
		_, _, err := net.SplitHostPort(peer)
		if err != nil {
			http.Error(w, fmt.Sprintf(`bad peer "host:port" format: %s`, err), http.StatusBadRequest)
			return
		}
	}

	peersMutex.Lock()
	defer peersMutex.Unlock()
	for _, peer := range newPeers {
		if _, present := peers[peer]; !present {
			if *verbose {
				log.Printf("Added peer %q.", peer)
			}
			peers[peer] = struct{}{}
		}
	}
}
