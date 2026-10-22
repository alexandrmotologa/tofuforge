package server

import (
	"io/fs"

	"github.com/alexandrmotologa/tofuforge/internal/assets"
)

// DistFS references the web studio compiled frontend assets.
var DistFS fs.FS = assets.DistFS
