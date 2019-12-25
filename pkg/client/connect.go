package client

import (
	"fmt"

	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type connecting struct{}
type requireAuth struct {
	method string
	action func() error
}

func (c connecting) onPacket(h Handler, pkt packet.Packet) (flow, error) {
	connect, ok := pkt.(*packet.Connect)
	if !ok {
		return nil, fmt.Errorf("TODO: expect connect")
	}

	// TODO: refactor
	action := func() error {
		assignedID, sessionPresent, err := h.performConnect(connect.Payload().ClientID(), connect.CleanStart())
		if err != nil {
			h.connackWithError(err)
			return nil
		}

		h.connackWithSuccess(assignedID, sessionPresent)
		return nil
	}
	prop, ok := connect.Props()[packet.AuthenticationMethod]
	if ok {
		authMethod := prop[0].Payload().(packet.StringPropPayload)
		return requireAuth{
			method: string(authMethod),
			action: action,
		}, nil
	}

	err := action()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (c connecting) onError(h Handler, err error) flow {
	h.parent.Error(err)

	h.connackWithError(fmt.Errorf("received error in connect flow: %v", err))
	return nil
}

func (a requireAuth) onPacket(h Handler, pkt packet.Packet) (flow, error) {
	h.log.Debug("authenticating")
	a.action()
	return nil, nil
}

func (a requireAuth) onError(h Handler, err error) flow {
	h.log.Debug("error @ onError", zap.Error(err))
	return nil
}
