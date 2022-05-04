// Package common implements the utility functions shared across repositories.
package common

import (
	"strings"
	"testing"
)

func TestRandName(t *testing.T) {
	type args struct {
		words   int
		wordlen int
		delim   string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "single word name without delimiter",
			args: args{
				words:   1,
				wordlen: 10,
				delim:   "+",
			},
		},
		{
			name: "3 word name with delimiter",
			args: args{
				words:   3,
				wordlen: 5,
				delim:   "-",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RandName(tt.args.words, tt.args.wordlen, tt.args.delim)

			if len(got) != tt.args.words*tt.args.wordlen+len(tt.args.delim)*(tt.args.words-1) ||
				strings.Count(got, tt.args.delim) < (tt.args.words-1) {
				t.Errorf("%s failed, args %#v, got %s", tt.name, tt.args, got)
			}
		})
	}
}
