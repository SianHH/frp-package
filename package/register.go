package service

import (
	"errors"
	"sync"
)

type Service interface {
	Start() error
	Stop()
	Wait()
}

var register = sync.Map{}

func Set(name string, svc Service) error {
	if _, ok := register.Load(name); ok {
		return errors.New("already exists")
	}
	register.Store(name, svc)
	return nil
}

func Get(name string) Service {
	value, ok := register.Load(name)
	if ok {
		return value.(Service)
	}
	return nil
}

func Del(name string) {
	value, ok := register.Load(name)
	if ok {
		value.(Service).Stop()
		register.Delete(name)
	}
}

func Range(f func(key string, value Service)) {
	var tempMap = make(map[string]Service)
	register.Range(func(key, value any) bool {
		service := value.(Service)
		tempMap[key.(string)] = service
		return true
	})
	for k, v := range tempMap {
		f(k, v)
	}
}
