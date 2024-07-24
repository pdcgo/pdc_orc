package file_version_test

import (
	"testing"

	"github.com/pdcgo/pdc_orc/pkg/file_version"
	"github.com/stretchr/testify/assert"
)

func TestFileVersion(t *testing.T) {
	versi := file_version.NewFileVersion("/tmp/sdk.db")
	err := versi.CreateVersion("pertama")
	assert.Nil(t, err)

	err = versi.CopyVersionFrom("pertama", "kedua")
	assert.Nil(t, err)

	err = versi.ActivateVersion("kedua")
	assert.Nil(t, err)
}
