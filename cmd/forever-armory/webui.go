package main

import "embed"

// webUI is the committed Vite build output. Runtime never builds the frontend.
//
//go:embed all:webui
var webUI embed.FS
