//go:generate npm run tss

// Package with static files
package static

import (
	"embed"
	"os"
)

//go:embed css/* img/* fonts/* js/*
var _StaticFiles embed.FS

var StaticFiles = os.DirFS("./modules/web/web_src")