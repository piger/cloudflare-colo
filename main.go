package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// {"iata":"TIA","lat":41.4146995544,"lon":19.7206001282,"cca2":"AL","region":"Europe","city":"Tirana"}
type ColoLocation struct {
	Iata   string
	Lat    float64
	Lon    float64
	CCA2   string
	Region string
	City   string
}

type Colo struct {
	Name      string
	Iata      string
	Continent string
	ColoLocation
}

// Antananarivo, Madagascar - (TNR)
var splitColoRe = regexp.MustCompile(`(.*?) - \(([^)]+)\)$`)

func splitColoString(s string) (string, string) {
	m := splitColoRe.FindStringSubmatch(s)
	if len(m) != 3 {
		return "", ""
	}

	return m[1], m[2]
}

func parseLocations() ([]ColoLocation, error) {
	fh, err := os.Open("locations.json")
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	var locations []ColoLocation
	if err := json.NewDecoder(fh).Decode(&locations); err != nil {
		return nil, err
	}

	return locations, nil
}

func run() error {
	/*
		site := "https://www.cloudflarestatus.com/"

		resp, err := http.Get(site)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	*/
	fh, err := os.Open("status.html")
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(fh)
	if err != nil {
		return err
	}

	colos := make(map[string]Colo)

	doc.Find("div.component-container").Each(func(i int, s *goquery.Selection) {
		continent := strings.TrimSpace(
			s.Find(`div.component-inner-container > span.name > span:not([class~="font-small"])`).Text(),
		)
		if continent == "Cloudflare Sites and Services" {
			return
		}

		fmt.Printf("# %s\n", continent)

		s.Find("div.child-components-container > div.component-inner-container > span.name").Each(func(i int, s *goquery.Selection) {
			colo := strings.TrimSpace(s.Text())
			where, iata := splitColoString(colo)
			fmt.Printf("- colo=%q where=%q, iata=%q\n", colo, where, iata)
			colos[iata] = Colo{Name: where, Iata: iata, Continent: continent}
		})
	})

	locations, err := parseLocations()
	if err != nil {
		return err
	}

	for _, location := range locations {
		c, ok := colos[location.Iata]
		if ok {
			c.Lat = location.Lat
			c.Lon = location.Lon
			c.CCA2 = location.CCA2
			c.Region = location.Region
			c.City = location.City
			colos[location.Iata] = c
		}
	}

	for key, value := range colos {
		fmt.Printf("colo: %+v (key=%q)\n", value, key)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
