package frpc

import (
	"context"
	"github.com/SianHH/frp-package/client"
	v1 "github.com/SianHH/frp-package/pkg/config/v1"
	"github.com/SianHH/frp-package/pkg/config/v1/validation"
	"github.com/SianHH/frp-package/pkg/featuregate"
	"time"
)

type Option func(s *Service) error

func FromBytes(data []byte) Option {
	return func(s *Service) error {
		// 不启用严格模式，尽量适配配置内容
		cfg, proxyCfgs, visitorCfgs, _, err := LoadClientConfig(data, false)
		if err != nil {
			return err
		}
		if len(cfg.FeatureGates) > 0 {
			if err := featuregate.SetFromMap(cfg.FeatureGates); err != nil {
				return err
			}
		}
		if _, err := validation.ValidateAllClientConfig(cfg, proxyCfgs, visitorCfgs); err != nil {
			return err
		}
		s.common = *cfg
		s.proxyCfgs = proxyCfgs
		s.visitorCfgs = visitorCfgs
		return nil
	}
}

type Service struct {
	common      v1.ClientCommonConfig
	proxyCfgs   []v1.ProxyConfigurer
	visitorCfgs []v1.VisitorConfigurer
	svc         *client.Service
	stopChan    chan struct{}
}

func NewService(common v1.ClientCommonConfig, proxyCfgs []v1.ProxyConfigurer, visitorCfgs []v1.VisitorConfigurer, opts ...Option) (*Service, error) {
	s := &Service{
		common:      common,
		proxyCfgs:   proxyCfgs,
		visitorCfgs: visitorCfgs,
		stopChan:    make(chan struct{}),
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	return s, nil
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
