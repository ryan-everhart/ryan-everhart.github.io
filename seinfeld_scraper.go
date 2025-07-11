package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Episode struct {
	Season      int    `json:"season"`
	Episode     int    `json:"episode"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func main() {
	var allEpisodes []Episode

	for season := 1; season <= 9; season++ {
		url := fmt.Sprintf("https://en.wikipedia.org/wiki/Seinfeld_season_%d", season)
		fmt.Printf("Scraping Season %d: %s\n", season, url)

		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("failed to get season %d: %v", season, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Fatalf("status code error for season %d: %d %s", season, resp.StatusCode, resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			log.Fatal(err)
		}

		episodeNum := 0

		doc.Find("table.wikiepisodetable tr").Each(func(i int, s *goquery.Selection) {
			if s.HasClass("vevent") {
				cells := s.Find("td")
				if cells.Length() >= 3 {
					episodeNum++
					title := strings.Trim(cells.Eq(1).Text(), "\" \n")
					allEpisodes = append(allEpisodes, Episode{
						Season:      season,
						Episode:     episodeNum,
						Title:       title,
						Description: "",
					})
				}
			}

			if s.HasClass("expand-child") {
				description := s.Find("div.shortSummaryText").Text()
				description = strings.TrimSpace(description)
				if len(allEpisodes) > 0 {
					allEpisodes[len(allEpisodes)-1].Description = description
				}
			}
		})
	}

	// Write to JSON file
	output, err := json.MarshalIndent(allEpisodes, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("seinfeld_episodes.json", output, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Data written to seinfeld_episodes.json")
}
