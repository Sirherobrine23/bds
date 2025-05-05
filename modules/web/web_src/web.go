//go:generate npm run tss

// Package with static files
package web

import "embed"

//go:embed css/* img/* fonts/* js/*
var StatisFiles embed.FS
