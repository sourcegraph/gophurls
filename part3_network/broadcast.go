package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// broadcastLink adds link to each peer.
func broadcastLink(link *link) error {
	// Serialize the link to JSON.
	var bodyBuf bytes.Buffer
	err := json.NewEncoder(&bodyBuf).Encode(link)
	if err != nil {
		return err
	}

	peersMutex.RLock()
	defer peersMutex.RUnlock()
	for peer, _ := range peers {
		if *verbose {
			log.Printf("Broadcasting to peer %q: %v", peer, link)
		}
		resp, err := http.Post(fmt.Sprintf("http://%s/links", peer), "application/json", &bodyBuf)
		if err != nil {
			return fmt.Errorf("broadcasting to peer %q: %s", peer, err)
		}
		resp.Body.Close()
		if *verbose {
			log.Printf("Finished broadcasting to peer %q: %v", peer, link)
		}
	}

	return nil
}
