package main

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	statusPageURL    = "https://www.cloudflarestatus.com/"
	locationsPageURL = "https://speed.cloudflare.com/locations"

	userAgent = "my little scraper"
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
var splitColoRe = regexp.MustCompile(`(.*?)\s+-[^(]+\(([^)]+)\)$`)

func splitColoString(s string) (string, string) {
	m := splitColoRe.FindStringSubmatch(s)
	if len(m) != 3 {
		return "", ""
	}

	return strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
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
		switch continent {
		case "Cloudflare Sites and Services":
			return
		case "":
			err = errors.Join(err, errors.New("empty continent"))
		}

		s.Find("div.child-components-container > div.component-inner-container > span.name").Each(func(i int, s *goquery.Selection) {
			coloText := strings.TrimSpace(s.Text())

			where, iata := splitColoString(coloText)
			if where == "" || iata == "" {
				err = errors.Join(err, fmt.Errorf("error extracting colo data from %s", coloText))
			}

			m[iata] = Colo{Name: where, Iata: iata, Continent: continent}
		})
	})

	return m, err
}

func parseLocationsJSON(r io.Reader) ([]Location, error) {
	var locations []Location
	if err := json.NewDecoder(r).Decode(&locations); err != nil {
		return locations, err
	}

	return locations, nil
}

func enrichColoMap(m map[string]Colo, locations []Location) {
	for _, location := range locations {
		c, ok := m[location.Iata]
		if ok {
			c.Lat = location.Lat
			c.Lon = location.Lon
			c.CCA2 = location.CCA2
			c.Region = location.Region
			c.City = location.City
			m[location.Iata] = c
		}
	}
}

func fetchPage(ctx context.Context, client *http.Client, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code while fetching URL %s: %d", url, resp.StatusCode)
	}

	return resp.Body, nil
}

func getColoMap(ctx context.Context, client *http.Client) (map[string]Colo, error) {
	body, err := fetchPage(ctx, client, statusPageURL)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	coloMap, err := parseStatusPage(body)
	return coloMap, err
}

func getLocations(ctx context.Context, client *http.Client) ([]Location, error) {
	body, err := fetchPage(ctx, client, locationsPageURL)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	locations, err := parseLocationsJSON(body)
	return locations, err
}

func run(filename string) error {
	ctx := context.Background()
	client := &http.Client{}

	coloMap, err := getColoMap(ctx, client)
	if err != nil {
		return err
	}

	locations, err := getLocations(ctx, client)
	if err != nil {
		return err
	}

	enrichColoMap(coloMap, locations)

	coloList := sortColos(coloMap)

	if err := MarshalColos(coloList, filename); err != nil {
		return err
	}

	return nil
}

func main() {
	var outputFilename string
	flag.StringVar(&outputFilename, "output", "colos.json", "Name of the file where to write the colo map")
	flag.Parse()

	if err := run(outputFilename); err != nil {
		log.Fatal(err)
	}
}
