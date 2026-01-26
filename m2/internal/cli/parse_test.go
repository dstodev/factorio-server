package cli_test

import (
	"manage2/internal/cli"
	"manage2/internal/test"
	"testing"
)

func TestSplitNextDelimiter(t *testing.T) {
	testCases := []struct {
		name            string
		input           []string
		expectNext      []string
		expectRemaining []string
	}{
		{
			name:            "No Args",
			input:           []string{},
			expectNext:      []string{},
			expectRemaining: nil,
		},
		{
			name:            "No Delimiter",
			input:           []string{"some", "--arg", "-args"},
			expectNext:      []string{"some", "--arg", "-args", "--"},
			expectRemaining: nil,
		},
		{
			name:            "With Delimiter",
			input:           []string{"some", "--arg", "-args", "--"},
			expectNext:      []string{"some", "--arg", "-args", "--"},
			expectRemaining: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			next, remaining := cli.SplitNextDelimiter(tc.input)
			test.AssertEqual(t, tc.expectNext, next)
			test.AssertEqual(t, tc.expectRemaining, remaining)
		})
	}
}

func TestDelimit(t *testing.T) {
	testCases := []struct {
		name   string
		input  []string
		expect []string
	}{
		{
			name:   "No Args",
			input:  []string{},
			expect: []string{},
		},
		{
			name:   "Empty String",
			input:  []string{""},
			expect: []string{"", "--"},
		},
		{
			name:   "Positional",
			input:  []string{"a"},
			expect: []string{"a", "--"},
		},
		{
			name:   "One Arg",
			input:  []string{"-a"},
			expect: []string{"-a", "--"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := cli.Delimit(tc.input)
			test.AssertEqual(t, tc.expect, opts)
		})
	}
}

func TestExpand(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expect []string
	}{
		{
			name:   "Empty String",
			input:  "",
			expect: []string{""},
		},
		{
			name:   "Positional",
			input:  "a",
			expect: []string{"a"},
		},
		{
			name:   "One Arg",
			input:  "-a",
			expect: []string{"-a"},
		},
		{
			name:   "Two Args",
			input:  "-ab",
			expect: []string{"-a", "-b"},
		},
		{
			name:   "Ignore Long Args",
			input:  "--input",
			expect: []string{"--input"},
		},
		{
			name:   "Space in input",
			input:  "-a b",
			expect: []string{"-a", "- ", "-b"},
		},
		{
			name:   "Dash in input",
			input:  "-a-b",
			expect: []string{"-a", "--", "-b"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := cli.Expand(tc.input)
			test.AssertEqual(t, tc.expect, opts)
		})
	}
}
