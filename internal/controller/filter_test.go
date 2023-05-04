package controller_test

import (
	"github.com/garethjevans/filter-controller/internal/controller"
	"github.com/stretchr/testify/assert"
	"golang.org/x/mod/sumdb/dirhash"
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

	filtered := controller.FilterFileList(files, include)

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
