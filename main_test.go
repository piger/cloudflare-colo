package main

import (
	"cmp"
	"context"
	"flag"
	"net/http"
	"os"
	"testing"
)

var useLiveData bool

func TestSplitColoString(t *testing.T) {
	tests := []struct {
		Input    string
		Name     string
		Iata     string
		Expected []string
	}{
		{
			Input: "Johor Bahru, Malaysia - (JHB)",
			Name:  "Johor Bahru, Malaysia",
			Iata:  "JHB",
		},
		{
			Input: "Cork, Ireland -  (ORK)",
			Name:  "Cork, Ireland",
			Iata:  "ORK",
		},
		{
			Input: "Milan, Italy - (MXP)",
			Name:  "Milan, Italy",
			Iata:  "MXP",
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			name, iata := splitColoString(test.Input)
			if cmp.Compare(test.Name, name) != 0 {
				t.Fatalf("expected %q, got %q", test.Name, name)
			}

			if cmp.Compare(test.Iata, iata) != 0 {
				t.Fatalf("expected %q, got %q", test.Iata, iata)
			}
		})
	}
}

func TestParseStatusPageLive(t *testing.T) {
	if !useLiveData {
		t.Skip()
	}

	tests := []struct {
		Iata  string
		Name  string
		Group string
	}{
		{
			Iata:  "BHY",
			Name:  "Beihai, China",
			Group: "Asia",
		},
	}

	ctx := context.Background()
	client := &http.Client{}
	coloMap, err := getColoMap(ctx, client)
	if err != nil {
		t.Fatalf("failed to fetch colo map: %s", err)
	}

	for _, test := range tests {
		t.Run(test.Iata, func(t *testing.T) {
			colo, ok := coloMap[test.Iata]
			if !ok {
				t.Fatalf("colo %s was not found in map", test.Iata)
			}

			if cmp.Compare(test.Name, colo.Name) != 0 {
				t.Fatalf("colo %s expected name %s, got %s", test.Iata, test.Name, colo.Name)
			}

			if cmp.Compare(test.Group, colo.Group) != 0 {
				t.Fatalf("colo %s expected group %s, got %s", test.Iata, test.Group, colo.Group)
			}
		})
	}
}

func TestMain(m *testing.M) {
	flag.BoolVar(&useLiveData, "live", false, "Use live data in tests")
	flag.Parse()

	os.Exit(m.Run())
}
