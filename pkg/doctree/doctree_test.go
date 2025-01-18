package doctree_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/WinPooh32/go-coder/pkg/doctree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustReadFile(name string) string {
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		panic(err)
	}

	return string(data)
}

func TestNewDocument(t *testing.T) {
	t.Parallel()

	type args struct {
		url string
	}

	tests := []struct {
		name string
		args args
		want *doctree.Document
	}{
		{
			name: "Test with local file",
			args: args{
				url: filepath.Join("testdata", "sample.md"),
			},
			want: &doctree.Document{
				URL:     filepath.Join("testdata", "sample.md"),
				Content: mustReadFile("sample.md"),
				Links:   []string{filepath.Join("testdata", "another-document.md")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := doctree.NewDocument(tt.args.url)
			require.NoError(t, err)

			// Compare URLs and Content directly
			assert.Equal(t, tt.want.URL, got.URL)
			assert.Equal(t, tt.want.Content, got.Content)

			// Compare Links separately to handle relative paths correctly
			if len(got.Links) != len(tt.want.Links) {
				t.Errorf("got %d links, want %d", len(got.Links), len(tt.want.Links))
			} else {
				for i := range got.Links {
					assert.Equal(t, tt.want.Links[i], got.Links[i])
				}
			}

			// Since LinkedBy is nil in the expected value and might be populated later,
			// we check if it's not nil only if the test case expects it to be.
			if len(tt.want.LinkedBy) != 0 {
				assert.NotNil(t, got.LinkedBy)
			} else {
				assert.Nil(t, got.LinkedBy)
			}
		})
	}
}

func TestBuildGraph(t *testing.T) {
	t.Parallel()

	type args struct {
		urls []string
	}

	tests := []struct {
		name string
		args args
		want map[string]*doctree.Document
	}{
		{
			name: "Test with multiple local files",
			args: args{
				urls: []string{
					filepath.Join("testdata", "sample.md"),
					filepath.Join("testdata", "another-document.md"),
				},
			},
			want: map[string]*doctree.Document{
				filepath.Join("testdata", "sample.md"): {
					URL:     filepath.Join("testdata", "sample.md"),
					Content: mustReadFile("sample.md"),
					Links:   []string{filepath.Join("testdata", "another-document.md")},
				},
				filepath.Join("testdata", "another-document.md"): {
					URL:     filepath.Join("testdata", "another-document.md"),
					Content: "# Another Document\n\nThis is another document.\n",
					Links:   []string{},
					LinkedBy: []*doctree.Document{
						{URL: filepath.Join("testdata", "sample.md")},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := doctree.BuildGraph(tt.args.urls)
			require.NoError(t, err)

			for k, v := range got {
				wantDoc, exists := tt.want[k]
				if !exists {
					t.Errorf("unexpected document: %s", k)
					continue
				}

				assert.Equal(t, wantDoc.URL, v.URL)
				assert.Equal(t, wantDoc.Content, v.Content)

				if len(v.Links) != len(wantDoc.Links) {
					t.Errorf("got %d links for %s, want %d", len(v.Links), k, len(wantDoc.Links))
				} else {
					for i := range v.Links {
						assert.Equal(t, wantDoc.Links[i], v.Links[i])
					}
				}

				if len(v.LinkedBy) != len(wantDoc.LinkedBy) {
					t.Errorf("got %d LinkedBy for %s, want %d", len(v.LinkedBy), k, len(wantDoc.LinkedBy))
				} else {
					for i := range v.LinkedBy {
						assert.Equal(t, wantDoc.LinkedBy[i].URL, v.LinkedBy[i].URL)
					}
				}
			}
		})
	}
}

func TestFormatGraph(t *testing.T) {
	t.Parallel()

	type args struct {
		documents map[string]*doctree.Document
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test with multiple local files",
			args: args{
				documents: map[string]*doctree.Document{
					filepath.Join("testdata", "sample.md"): {
						URL:     filepath.Join("testdata", "sample.md"),
						Content: mustReadFile("sample.md"),
						Links:   []string{filepath.Join("testdata", "another-document.md")},
					},
					filepath.Join("testdata", "another-document.md"): {
						URL:     filepath.Join("testdata", "another-document.md"),
						Content: mustReadFile("another-document.md"),
						Links:   []string{},
						LinkedBy: []*doctree.Document{
							{URL: filepath.Join("testdata", "sample.md")},
						},
					},
				},
			},
			want: `Document: testdata/another-document.md
Linked By:
  - testdata/sample.md

Document: testdata/sample.md
Linked By:

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := doctree.FormatGraph(tt.args.documents)
			assert.Equal(t, tt.want, got)
		})
	}
}
