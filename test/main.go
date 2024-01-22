package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	count := 0
	for {
		count++
		time.Sleep(1 * time.Second)
		start := time.Now()
		httpClient := new(http.Client)
		httpClient.Timeout = 10 * time.Second
		resp, err := httpClient.Get("http://ipinfo.io")
		if err == nil {
			rb, err := io.ReadAll(resp.Body)
			if err == nil {
				log.Printf("(%d/%d/%d)\n", count, time.Since(start).Milliseconds(), len(rb))
			}
		} else {
			log.Println("ERROR!", err)
		}
	}
}
