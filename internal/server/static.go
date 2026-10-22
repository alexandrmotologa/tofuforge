package server

import "embed"

// DistFS embeds the web studio compiled frontend assets.
//
//go:embed all:dist
var DistFS embed.FS
