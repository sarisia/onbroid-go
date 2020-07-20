package main

import "testing"

func TestGetMessageID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"good url", "https://discordapp.com/channels/500660597310881802/500660597310881804/734521106152554527", "734521106152554527"},
		{"good id", "734521106152554527", "734521106152554527"},
		{"bad url", "https://discordapp.com/channels/invalid1223334444", ""},
		{"bad id", "invalid1223334444", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getMessageID(tc.input)
			if got != tc.want {
				t.Errorf("want: %s, got: %s\n", tc.want, got)
			}
		})
	}
}

func TestGetChannelID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"good id (reply style)", "<#572474514903007272>", "572474514903007272"},
		{"good id (string style)", "572474514903007272", "572474514903007272"},
		{"bad id", "badid", ""},
		{"bad id with digits", "badid1223334444", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := getChannelID(tc.input)
			if got != tc.want {
				t.Errorf("want: %s, got: %s\n", tc.want, got)
			}
		})
	}
}
