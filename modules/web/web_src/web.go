//go:generate npm run tss

// Package with static files
package static

import "embed"

//go:embed css/* img/* fonts/* js/*
var StaticFiles embed.FS
