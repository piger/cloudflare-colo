package main

import (
	"cmp"
	"encoding/json"
	"io"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Location struct {
	Iata   string  `json:"iata"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	CCA2   string  `json:"cca2"`
	Region string  `json:"region"`
	City   string  `json:"city"`
}

type Colo struct {
	Name      string `json:"name"`
	Continent string `json:"continent"`
	Iata      string `json:"iata"`

	Lat    float64 `json:"lat,omitempty"`
	Lon    float64 `json:"lon,omitempty"`
	CCA2   string  `json:"cca2,omitempty"`
	Region string  `json:"region,omitempty"`
	City   string  `json:"city,omitempty"`
}

func MarshalColos(colos []Colo, filename string) error {
	fh, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fh.Close()

	enc := json.NewEncoder(fh)
	enc.SetIndent("", "  ")
	return enc.Encode(colos)
}

func sortColos(colos map[string]Colo) []Colo {
	var rv []Colo
	for _, colo := range colos {
		rv = append(rv, colo)
	}

	slices.SortFunc(rv, func(a, b Colo) int {
		x := cmp.Compare(a.Continent, b.Continent)
		if x == 0 {
			return cmp.Compare(a.Name, b.Name)
		}
		return x
	})

	return rv
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

func parseStatusPage(r io.Reader) (map[string]Colo, error) {
	m := make(map[string]Colo)

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return m, err
	}

	doc.Find("div.component-container").Each(func(i int, s *goquery.Selection) {
		continent := strings.TrimSpace(
			s.Find(`div.component-inner-container > span.name > span:not([class~="font-small"])`).Text(),
		)
		if continent == "Cloudflare Sites and Services" {
			return
		}

		s.Find("div.child-components-container > div.component-inner-container > span.name").Each(func(i int, s *goquery.Selection) {
			where, iata := splitColoString(strings.TrimSpace(s.Text()))
			m[iata] = Colo{Name: where, Iata: iata, Continent: continent}
		})
	})

	return m, nil
}

func parseLocationsJSON(r io.Reader) ([]Location, error) {
	var locations []Location
	if err := json.NewDecoder(r).Decode(&locations); err != nil {
		return locations, err
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
	defer fh.Close()

	coloMap, err := parseStatusPage(fh)
	if err != nil {
		return err
	}

	fh2, err := os.Open("locations.json")
	if err != nil {
		return err
	}
	defer fh2.Close()

	locations, err := parseLocationsJSON(fh2)
	if err != nil {
		return err
	}

	for _, location := range locations {
		c, ok := coloMap[location.Iata]
		if ok {
			c.Lat = location.Lat
			c.Lon = location.Lon
			c.CCA2 = location.CCA2
			c.Region = location.Region
			c.City = location.City
			coloMap[location.Iata] = c
		}
	}

	coloList := sortColos(coloMap)

	if err := MarshalColos(coloList, "colos.json"); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
