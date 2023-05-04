package controller_test

import (
	"github.com/fluxcd/go-git/v5/plumbing/format/gitignore"
	"github.com/fluxcd/pkg/sourceignore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/mod/sumdb/dirhash"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilter(t *testing.T) {
	files, err := dirhash.DirFiles("../..", ".")
	assert.NoError(t, err)

	t.Logf("Got files %s", files)

	include := `!.git
go.*
internal/**/*.go
api
!.*
!**/*_test.go
!**/tests`

	filtered, err := FilterFileList(files, include)
	assert.NoError(t, err)

	t.Logf("Filtered files %s", filtered)

	assert.True(t, len(filtered) > 0)

	assert.Contains(t, filtered, "go.mod")
	assert.Contains(t, filtered, "go.sum")
	assert.Contains(t, filtered, "internal/controller/filter_controller.go")
	assert.Contains(t, filtered, "api/v1alpha1/filter_types.go")

	assert.NotContains(t, filtered, ".gitignore")
	assert.NotContains(t, filtered, ".github/dependabot.yml")
	assert.NotContains(t, filtered, ".idea/.gitignore")

}

func FilterFileList(list []string, include string) ([]string, error) {
	var domain []string
	patterns := sourceignore.ReadPatterns(strings.NewReader(include), domain)
	matcher := sourceignore.NewDefaultMatcher(patterns, domain)

	logrus.Infof("got patterns %+v", patterns)

	var filtered []string
	for _, file := range list {
		logrus.Debugf("checking %s", file)

		fileParts := strings.Split(file, string(filepath.Separator))

		if matcher.Match(fileParts, false) {

			//if ignore.Ignore(file) {
			filtered = append(filtered, file)
		}
		//}
	}

	return filtered, nil
}

// ArchiveFileFilter must return true if a file should not be included in the archive after inspecting the given path
// and/or os.FileInfo.
type ArchiveFileFilter func(p string, fi os.FileInfo) bool

// SourceIgnoreFilter returns an ArchiveFileFilter that filters out files matching sourceignore.VCSPatterns and any of
// the provided patterns.
// If an empty gitignore.Pattern slice is given, the matcher is set to sourceignore.NewDefaultMatcher.
func SourceIgnoreFilter(ps []gitignore.Pattern, domain []string) ArchiveFileFilter {
	matcher := sourceignore.NewDefaultMatcher(ps, domain)
	if len(ps) > 0 {
		ps = append(sourceignore.VCSPatterns(domain), ps...)
		matcher = sourceignore.NewMatcher(ps)
	}
	return func(p string, fi os.FileInfo) bool {
		return matcher.Match(strings.Split(p, string(filepath.Separator)), fi.IsDir())
	}
}
