package main

import (
	"cmp"
	"testing"
)

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
