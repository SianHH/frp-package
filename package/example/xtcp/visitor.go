package main

import (
	"github.com/SianHH/frp-package/package/frpc"
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
	}, []v1.ProxyConfigurer{}, []v1.VisitorConfigurer{
		&v1.XTCPVisitorConfig{
			VisitorBaseConfig: v1.VisitorBaseConfig{
				Name: "sdfsdf",
				Type: "xtcp",
				Transport: v1.VisitorTransport{
					UseEncryption:  true,
					UseCompression: true,
				},
				SecretKey:  "******",
				ServerUser: "",
				ServerName: "111",
				BindAddr:   "0.0.0.0",
				BindPort:   22714,
				Plugin:     v1.TypedVisitorPluginOptions{},
			},
			Protocol:          "",
			KeepTunnelOpen:    true, // 保存打开
			MaxRetriesAnHour:  0,    // 每小时尝试打洞次数
			MinRetryInterval:  0,    // 每次打洞最小间隔
			FallbackTo:        "",   // 回落的STCP服务
			FallbackTimeoutMs: 0,    // 超时时间
		},
	})

	if err := svc.Start(); err != nil {
		log.Fatalln(err)
	}
	svc.Wait()
}
