package main

import (
	"github.com/SianHH/frp-package/package/frpc"
	"github.com/SianHH/frp-package/pkg/config/types"
	v1 "github.com/SianHH/frp-package/pkg/config/v1"
	"log"
)

func main() {
	svc := frpc.NewService(v1.ClientCommonConfig{
		Auth: v1.AuthClientConfig{
			Token: "123123123",
		},
		User:       "",
		ServerAddr: "127.0.0.1",
		ServerPort: 7000,
		Transport: v1.ClientTransportConfig{
			Protocol: "tcp",
		},
	}, []v1.ProxyConfigurer{
		&v1.XTCPProxyConfig{
			ProxyBaseConfig: v1.ProxyBaseConfig{
				Name:        "111",
				Type:        "xtcp",
				Annotations: nil,
				Transport: v1.ProxyTransport{
					UseEncryption:  true,
					UseCompression: true,
					BandwidthLimit: func() types.BandwidthQuantity {
						quantity, _ := types.NewBandwidthQuantity("128KB")
						return quantity
					}(),
					BandwidthLimitMode:   "client",
					ProxyProtocolVersion: "",
				},
				Metadatas:    nil,
				LoadBalancer: v1.LoadBalancerConfig{},
				HealthCheck:  v1.HealthCheckConfig{},
				ProxyBackend: v1.ProxyBackend{
					LocalIP:   "192.168.0.172",
					LocalPort: 22714,
					Plugin:    v1.TypedClientPluginOptions{},
				},
			},
			Secretkey:  "******",
			AllowUsers: []string{"*"},
		},
	}, []v1.VisitorConfigurer{})

	if err := svc.Start(); err != nil {
		log.Fatalln(err)
	}
	svc.Wait()
}
