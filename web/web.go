package web

import "embed"

//go:embed public
var PublicBox embed.FS

//go:embed template
var TemplateBox embed.FS
