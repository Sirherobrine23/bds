//go:generate npm run webpacked
package web


import (
	"embed"
)

var (
	//go:embed css/** js/** img/**
	StatisFiles embed.FS
)
