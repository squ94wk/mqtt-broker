package client

import (
	"fmt"

	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type initial struct{}
type connectReqAuth struct {
	method string
	action func() error
}

func (c initial) onPacket(client *Client, pkt packet.Packet) (state, error) {
	connect, ok := pkt.(*packet.Connect)
	if !ok {
		return nil, fmt.Errorf("TODO: expect connect")
	}

	// TODO: refactor
	action := func() error {
		assignedID, sessionPresent, err := client.parent.PerformConnect(connect.Payload().ClientID(), connect.CleanStart(), client)
		if err != nil {
			client.connackWithError(err)
			return nil
		}

		client.clientID = assignedID
		client.connackWithSuccess(assignedID, sessionPresent)
		return nil
	}
	prop, ok := connect.Props()[packet.AuthenticationMethod]
	if ok {
		authMethod := prop[0].Payload().(packet.StringPropPayload)
		return connectReqAuth{
			method: string(authMethod),
			action: action,
		}, nil
	}

	err := action()
	if err != nil {
		return nil, err
	}
	return connected{}, nil
}

func (c initial) onError(client *Client, err error) state {
	client.parent.Error(err)

	client.connackWithError(fmt.Errorf("received error in connect state: %v", err))
	return nil
}

func (a connectReqAuth) onPacket(client *Client, pkt packet.Packet) (state, error) {
	client.log.Debug("authenticating")
	err := a.action()
	if err != nil {
		client.log.Error("failed to connect", zap.Error(err))
	}
	return connected{}, nil
}

func (a connectReqAuth) onError(client *Client, err error) state {
	client.log.Debug("error @ onError", zap.Error(err))
	return nil
}
