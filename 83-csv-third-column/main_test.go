package main

import (
	"slices"
	"strings"
	"testing"
)

func TestThirdColumnContent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name: "plain rows with header skipped",
			input: "name,age,city\n" +
				"Alice,30,Berlin\n" +
				"Bob,25,Lagos\n",
			want: []string{"Berlin", "Lagos"},
		},
		{
			name: "quoted comma is one field",
			// The reason strings.Split can never parse CSV.
			input: "name,age,city\n" +
				"Alice,30,Berlin\n" +
				"\"Doe, John\",41,\"Paris, FR\"\n",
			want: []string{"Berlin", "Paris, FR"},
		},
		{
			name: "escaped quotes inside quoted field",
			input: "h1,h2,h3\n" +
				"a,b,\"say \"\"hey\"\"\"\n",
			want: []string{`say "hey"`},
		},
		{
			name: "embedded newline inside quoted field",
			// A record is not a line: this is ONE record spanning two lines.
			input: "h1,h2,h3\n" +
				"a,b,\"line one\nline two\"\n",
			want: []string{"line one\nline two"},
		},
		{
			name: "short rows skipped silently",
			input: "h1,h2,h3\n" +
				"just,two\n" +
				"x,y,z\n" +
				"lonely\n" +
				"p,q,r\n",
			want: []string{"z", "r"},
		},
		{
			name: "wide rows use index 2",
			input: "h1,h2,h3\n" +
				"p,q,r,s,t\n",
			want: []string{"r"},
		},
		{
			name: "empty third field contributes empty string",
			input: "h1,h2,h3\n" +
				"a,b,\n" +
				",,\n",
			want: []string{"", ""},
		},
		{
			name: "whitespace kept verbatim",
			input: "h1,h2,h3\n" +
				"a,b,  spaced  \n",
			want: []string{"  spaced  "},
		},
		{
			name: "crlf endings",
			input: "name,age,city\r\n" +
				"Alice,30,Berlin\r\n" +
				"Bob,25,Lagos\r\n",
			want: []string{"Berlin", "Lagos"},
		},
		{
			name: "no trailing newline on last record",
			input: "h1,h2,h3\n" +
				"a,b,c",
			want: []string{"c"},
		},
		{
			name: "blank lines between records ignored",
			input: "h1,h2,h3\n" +
				"\n" +
				"a,b,c\n" +
				"\n" +
				"d,e,f\n",
			want: []string{"c", "f"},
		},
		{
			name: "header skipped even when short",
			input: "only,two\n" +
				"a,b,c\n",
			want: []string{"c"},
		},
		{
			name: "header skipped even when it looks like data",
			input: "a,b,c\n" +
				"d,e,f\n",
			want: []string{"f"},
		},
		{
			name:  "unicode fields",
			input: "h1,h2,h3\nAçaí,30,\"Reykjavík, IS\"\n",
			want:  []string{"Reykjavík, IS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ThirdColumn(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("ThirdColumn() error = %v, want nil", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("ThirdColumn() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestThirdColumnNothingToReturn(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty input", ""},
		{"header only", "name,age,city\n"},
		{"header only no trailing newline", "name,age,city"},
		{"header then blank lines", "name,age,city\n\n\n"},
		{"header then only short rows", "name,age,city\na,b\nc\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ThirdColumn(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("ThirdColumn() error = %v, want nil — nothing to return is not an error", err)
			}
			if len(got) != 0 {
				t.Errorf("ThirdColumn() = %q, want no results", got)
			}
		})
	}
}

func TestThirdColumnMalformed(t *testing.T) {
	// A bare quote in a non-quoted field is a parse error; it must reach
	// the caller, not be swallowed into an empty result.
	input := "h1,h2,h3\n" +
		"a,b\"broken,c\n"

	got, err := ThirdColumn(strings.NewReader(input))
	if err == nil {
		t.Fatalf("ThirdColumn(malformed) error = nil, want the parser's error surfaced")
	}
	if len(got) != 0 {
		t.Errorf("got %d results alongside an error, want none", len(got))
	}
}
