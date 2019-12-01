package action

import (
	"fmt"

	"github.com/squ94wk/mqtt-broker/pkg/session"
	"github.com/squ94wk/mqtt-common/pkg/packet"
	"go.uber.org/zap"
)

type parent interface {
	Error(error)
}

type client interface {
	Packets() <-chan packet.Packet
	Errors() <-chan error
	Deliver(packet.Packet)
	Close() error
}

type Handler struct {
	parent   parent
	client   client
	log      *zap.Logger
	shutdown chan struct{}
}

func NewHandler(parent parent, client client, log *zap.Logger) Handler {
	return Handler{
		parent: parent,
		client: client,
		log:    log,
	}
}

type state interface {
	onError(Handler, error) (state, error)
	onPacket(Handler, packet.Packet) (state, error)
}

func (h Handler) Start() {
	h.log.Info("start action handler")
	var state state = connecting{}
	for {
		select {
		case _ = <-h.shutdown:
			return

		case pkt := <-h.client.Packets():
			h.log.Debug("handle control packet")
			s, err := state.onPacket(h, pkt)
			if err != nil {
				h.closeWithError(packet.UnspecifiedError, fmt.Sprintf("server error"))
			}

			state = s

		case e := <-h.client.Errors():
			h.log.Debug("handle error")
			s, err := state.onError(h, e)
			if err != nil {
				h.closeWithError(packet.UnspecifiedError, fmt.Sprintf("server error"))
			}

			state = s
		}
	}
}

func (h Handler) Shutdown() {
	h.shutdown <- struct{}{}
}

type connecting struct{}
type requireAuth struct {
	method string
	action func() error
}
type operational struct{}
type closed struct{}

func (c connecting) onPacket(h Handler, pkt packet.Packet) (state, error) {
	connect, ok := pkt.(*packet.Connect)
	if !ok {
		return nil, fmt.Errorf("TODO: expect connect")
	}

	action := func() error {
		return h.performConnect(connect.Payload().ClientId(), connect.CleanStart())
	}
	prop, ok := connect.Props()[packet.AuthenticationMethod]
	if ok {
		authMethod := prop[0].(packet.StringProp)
		return requireAuth{
			method: authMethod.Value(),
			action: action,
		}, nil
	}

	action()
	return operational{}, nil
}

func (c connecting) onError(h Handler, err error) (state, error) {
	h.parent.Error(err)

	// TODO
	h.closeWithError(packet.UnspecifiedError, fmt.Sprintf("received error in connect state: %v", err))
	return operational{}, nil
}

func (a requireAuth) onPacket(h Handler, pkt packet.Packet) (state, error) {
	//auth, ok := pkt.(*packet.Auth)
	//if !ok {
	//return nil, fmt.Error("TODO: expect auth")
	//}

	// do auth stuff

	err := a.action()
	if err != nil {
		h.closeWithError(packet.UnspecifiedError, fmt.Sprintf("server failed to authenticate: %v", err))
	}
	return operational{}, nil
}

func (a requireAuth) onError(h Handler, err error) (state, error) {
	h.parent.Error(err)

	// TODO
	h.closeWithError(packet.UnspecifiedError, fmt.Sprintf("received error in auth state: %v", err))
	return operational{}, nil
}

func (o operational) onPacket(h Handler, pkt packet.Packet) (state, error) {
	switch pkt.(type) {
	//case *packet.Publish:
	//publish := pkt.(*packet.Publish)

	default:
		return nil, fmt.Errorf("TODO: expect auth")
	}
}

func (o operational) onError(h Handler, err error) (state, error) {
	h.parent.Error(err)

	// TODO
	h.closeWithError(packet.UnspecifiedError, fmt.Sprintf("received error in operational state: %v", err))
	return o, nil
}

func (c closed) onPacket(h Handler, pkt packet.Packet) (state, error) {
	return c, nil
}

func (c closed) onError(h Handler, err error) (state, error) {
	return c, nil
}

func (h Handler) performConnect(clientId string, cleanStart bool) error {
	h.log.Debug("perform connect", zap.String("clientId", clientId), zap.Bool("clean start", cleanStart))
	//TODO: refactor!!
	resp := make(chan func() (*session.Session, bool, error), 1)
	session.BindToSession(
		clientId,
		cleanStart,
		h.client,
		func(s *session.Session, sessionPresent bool, err error) {
			resp <- func() (*session.Session, bool, error) { return s, sessionPresent, err }
		})
	_, present, err := (<-resp)()
	if err != nil {
		return err
	}

	var connack packet.Connack
	connack.SetConnectReason(packet.Success)
	connack.SetSessionPresent(present)
	h.client.Deliver(connack)
	return nil
	//send connack
	//finish
	//TODO: lock session or client until connack is queued,
	//      so that it's the first thing the client gets no matter what

	// session (clean start or persisted)
	//   assigned client id

	// Connect Action
	// session takeover:
	//   will message
	//   disconnect packet
	// finish
	//   connack
	// post finish
	//   message delivery
	//   keep alive monitoring
}

func (h Handler) closeWithError(reasonCode packet.ConnectReason, reason string) {
	var connack packet.Connack
	connack.SetSessionPresent(false)
	connack.SetConnectReason(reasonCode)
	connack.ResetProps()
	connack.AddProp(packet.NewStringProp(packet.ReasonString, reason))
	h.client.Deliver(connack)
	err := h.client.Close()
	if err != nil {
		h.parent.Error(fmt.Errorf("failed to close client after error: %v", err))
	}

	h.Shutdown()
}
