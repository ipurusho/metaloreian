package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/imman/metaloreian/internal/scraper"
)

func main() {
	sc := scraper.NewClient()
	// Dump raw HTML of an album page
	html, err := sc.FetchHTMLPublic(context.Background(), "https://www.metal-archives.com/albums/Metallica/Master_of_Puppets/547")
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("/tmp/ma_album.html", []byte(html), 0644)
	fmt.Printf("wrote %d bytes to /tmp/ma_album.html\n", len(html))
}
