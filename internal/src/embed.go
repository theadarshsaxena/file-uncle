package src

import (
	"embed"
	_ "embed"
)

//go:embed html/serve.html
var ServeHTML string

//go:embed static/*
var StaticFiles embed.FS

//go:embed html/upload.html
var UploadHTML string