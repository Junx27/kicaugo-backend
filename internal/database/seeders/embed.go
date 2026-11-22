package seeders

import "embed"

// SeedFiles berisi file SQL seed yang di-embed ke binary.
//
//go:embed *.sql
var SeedFiles embed.FS
