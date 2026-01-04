package event

import (
	"errors"

	"github.com/SianHH/frp-package/pkg/msg"
)

var ErrPayloadType = errors.New("error payload type")

type Handler func(payload any) error

type StartProxyPayload struct {
	NewProxyMsg *msg.NewProxy
}

type CloseProxyPayload struct {
	CloseProxyMsg *msg.CloseProxy
}
