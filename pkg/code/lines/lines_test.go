package lines_test

import (
	"testing"

	"github.com/WinPooh32/go-coder/pkg/code/lines"
	"github.com/stretchr/testify/assert"
)

func TestAddNumbers(t *testing.T) {
	t.Parallel()

	type args struct {
		s string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{s: ""}, ""},
		{"one line", args{s: "package main"}, "1:package main\n"},
		{"one line with eol", args{s: "package main\n"}, "1:package main\n2:\n"},
		{"multiple lines", args{s: "package main\nfunc main() {\n}\n"}, "1:package main\n2:func main() {\n3:}\n4:\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, lines.AddNumbers(tt.args.s))
		})
	}
}
