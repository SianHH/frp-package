package frpc

import (
	"context"
	"github.com/SianHH/frp-package/client"
	v1 "github.com/SianHH/frp-package/pkg/config/v1"
	"time"
)

type Service struct {
	common      v1.ClientCommonConfig
	proxyCfgs   []v1.ProxyConfigurer
	visitorCfgs []v1.VisitorConfigurer
	svc         *client.Service
	stopChan    chan struct{}
}

func NewService(common v1.ClientCommonConfig, proxyCfgs []v1.ProxyConfigurer, visitorCfgs []v1.VisitorConfigurer) *Service {
	return &Service{
		common:      common,
		proxyCfgs:   proxyCfgs,
		visitorCfgs: visitorCfgs,
		stopChan:    make(chan struct{}),
	}
}

func (s *Service) Start() (err error) {
	s.common.Complete()
	for i := 0; i < len(s.proxyCfgs); i++ {
		s.proxyCfgs[i].Complete("")
	}
	for i := 0; i < len(s.visitorCfgs); i++ {
		s.visitorCfgs[i].Complete(&s.common)
	}
	if s.svc, err = client.NewService(client.ServiceOptions{
		Common:      &s.common,
		ProxyCfgs:   s.proxyCfgs,
		VisitorCfgs: s.visitorCfgs,
	}); err != nil {
		return err
	}
	go func() {
		err = s.svc.Run(context.Background())
	}()
	time.Sleep(time.Second)
	return err
}

func (s *Service) Stop() {
	s.svc.Close()
	close(s.stopChan)
}

func (s *Service) Wait() {
	select {
	case <-s.stopChan:
	}
}
