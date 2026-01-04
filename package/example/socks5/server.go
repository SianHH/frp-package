package main

import (
	"github.com/SianHH/frp-package/package/frps"
	v1 "github.com/SianHH/frp-package/pkg/config/v1"
	"log"
)

func main() {
	svc := frps.NewService(v1.ServerConfig{
		Auth: v1.AuthServerConfig{
			Token: "123123123",
		},
		BindAddr:      "0.0.0.0",
		BindPort:      7000,
		KCPBindPort:   0,
		QUICBindPort:  0,
		VhostHTTPPort: 18080,
		HTTPPlugins:   nil,
	})

	if err := svc.Start(); err != nil {
		log.Fatalln(err)
	}
	svc.Wait()
}
