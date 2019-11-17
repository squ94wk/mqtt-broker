package listener

import (
	"fmt"
	"net"
	"sync"

	"go.uber.org/zap"
)

type Parent interface {
	Error(error)
	OnNewConnection(net.Conn)
}

type Handler struct {
	parent   Parent
	log      *zap.Logger
	listener struct {
		sync.Mutex
		l net.Listener
	}
	shutdown chan struct{}
}

func NewHandler(parent Parent, log *zap.Logger) Handler {
	return Handler{
		parent: parent,
		log:    log,
		listener: struct {
			sync.Mutex
			l net.Listener
		}{},
	}
}

func (h *Handler) Start() {
	ln, err := net.Listen("tcp", "localhost:1883")
	if err != nil {
		h.parent.Error(fmt.Errorf("failed to start listener: failed to start listening on address %s: %v", "localhost:1883", err))
	}

	h.listener.Lock()
	h.listener.l = ln
	h.listener.Unlock()

	h.log.Info("accept connections")
	for {
		select {
		case _ = <-h.shutdown:
			h.log.Info("Listener stopped")
			return

		default:
			conn, err := ln.Accept()
			if err != nil {
				h.parent.Error(fmt.Errorf("failed to accept connections: %v", err))
				break
			}

			h.log.Info("new connection")
			h.parent.OnNewConnection(conn)
		}
	}
}

func (h *Handler) Shutdown() {
	select {
	case h.shutdown <- struct{}{}:
		break
	default:
		h.log.Info("listener handler has already received shutdown signal.")
	}

	h.listener.Lock()
	if h.listener.l != nil {
		err := h.listener.l.Close()
		if err != nil {
			h.parent.Error(fmt.Errorf("failed to close Listener: %v", err))
		}
	}
	h.listener.Unlock()
}
