package service

import (
	"errors"
	"sync"
)

type service interface {
	Start() error
	Stop()
	Wait()
}

var register = sync.Map{}

func Set(name string, svc service) error {
	if _, ok := register.Load(name); ok {
		return errors.New("already exists")
	}
	register.Store(name, svc)
	return nil
}

func Get(name string) service {
	value, ok := register.Load(name)
	if ok {
		return value.(service)
	}
	return nil
}

func Del(name string) {
	value, ok := register.Load(name)
	if ok {
		value.(service).Stop()
		register.Delete(name)
	}
}
