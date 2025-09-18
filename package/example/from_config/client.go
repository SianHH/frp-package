package main

import (
	"github.com/SianHH/frp-package/package/frpc"
	v1 "github.com/SianHH/frp-package/pkg/config/v1"
	"log"
)

func main() {
	svc, err := frpc.NewService(v1.ClientCommonConfig{}, nil, nil, frpc.FromBytes([]byte(`
serverAddr = "127.0.0.1"
serverPort = 7000

[[proxies]]
name = "test-tcp"
type = "tcp"
localIP = "127.0.0.1"
localPort = 28080
remotePort = 26000
`)))
	if err != nil {
		log.Fatalln(err)
	}
	if err := svc.Start(); err != nil {
		log.Fatalln(err)
	}
	svc.Wait()
}
