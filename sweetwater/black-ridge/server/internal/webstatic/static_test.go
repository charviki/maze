package webstatic

import (
	"io/fs"
	"testing"
)

func TestFiles_IsValidFS(t *testing.T) {
	var _ fs.FS = Files
}
