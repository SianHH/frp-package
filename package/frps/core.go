package frps

import (
	"context"
	v1 "github.com/SianHH/frp-package/pkg/config/v1"
	"github.com/SianHH/frp-package/server"
)

type Service struct {
	cfg      v1.ServerConfig
	svc      *server.Service
	stopChan chan struct{}
}

func NewService(cfg v1.ServerConfig) *Service {
	return &Service{
		cfg:      cfg,
		stopChan: make(chan struct{}),
	}
}

func (s *Service) Start() (err error) {
	s.cfg.Complete()
	s.svc, err = server.NewService(&s.cfg)
	if err != nil {
		return err
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
