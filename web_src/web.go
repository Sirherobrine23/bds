package web

import (
	"embed"
)

var (
	//go:embed css/** js/**
	StatisFiles embed.FS
)
