// Package web embeds the built Angular SPA so the server ships as a single
// self-contained binary. The Angular build outputs into dist/browser (see the
// frontend's angular.json outputPath).
package web

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:dist/browser
var dist embed.FS

// SPA returns the embedded build rooted at the browser output directory.
func SPA() (fs.FS, error) {
	sub, err := fs.Sub(dist, "dist/browser")
	if err != nil {
		return nil, fmt.Errorf("sub embedded spa fs: %w", err)
	}
	return sub, nil
}
