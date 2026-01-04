package main

import (
	"github.com/SianHH/frp-package/package/frps"
	v1 "github.com/SianHH/frp-package/pkg/config/v1"
	"log"
)

func main() {
	svc, err := frps.NewService(v1.ServerConfig{}, frps.FromBytes([]byte(`
bindPort = 7000
`)))
	if err != nil {
		log.Fatalln(err)
	}

	if err := svc.Start(); err != nil {
		log.Fatalln(err)
	}
	svc.Wait()
}
