package frpc

import (
	"embed"

	"github.com/SianHH/frp-package/assets"
)

//go:embed static/*
var content embed.FS

func init() {
	assets.Register(content)
}
