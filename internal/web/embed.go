// Package web embeds the four console HTML pages served by the envmonitor
// HTTP console.
package web

import "embed"

// Files holds the embedded HTML pages.
//
//go:embed *.html
var Files embed.FS
