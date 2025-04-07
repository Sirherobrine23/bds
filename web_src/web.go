//go:generate npm run tss

// Package with static files
package web

import "embed"

var (
	//go:embed css/* img/* fonts/* js/*
	StatisFiles embed.FS
)
