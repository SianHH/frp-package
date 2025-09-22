package frps

import (
	"context"
	v1 "github.com/SianHH/frp-package/pkg/config/v1"
	plugin "github.com/SianHH/frp-package/pkg/plugin/server"
	"github.com/SianHH/frp-package/server"
)

type Option func(s *Service) error

func FromBytes(data []byte) Option {
	return func(s *Service) error {
		svrCfg, _, err := LoadServerConfig(data, true)
		if err != nil {
			return err
		}
		s.cfg = *svrCfg
		return nil
	}
}

// 暴露注册插件方式
func RegistryPlugin(p plugin.Plugin) Option {
	return func(s *Service) error {
		if p == nil {
			return nil
		}
		s.plugins = append(s.plugins, p)
		return nil
	}
}

type Service struct {
	cfg      v1.ServerConfig
	svc      *server.Service
	stopChan chan struct{}
	plugins  []plugin.Plugin
}

func NewService(cfg v1.ServerConfig, opts ...Option) (*Service, error) {
	s := &Service{
		cfg:      cfg,
		stopChan: make(chan struct{}),
	}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Service) Start() (err error) {
	s.cfg.Complete()
	s.svc, err = server.NewService(&s.cfg)
	if err != nil {
		return err
	}
	manager := s.svc.GetPluginManager()
	for _, p := range s.plugins {
		manager.Register(p)
	}
	go s.svc.Run(context.Background())
	return nil
}

func (s *Service) Stop() {
	_ = s.svc.Close()
	close(s.stopChan)
}

func (s *Service) Wait() {
	select {
	case <-s.stopChan:
	}
}
