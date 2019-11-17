package listener

import (
	"fmt"
	"net"
	"sync"

	"go.uber.org/zap"
)

type Config interface {
	ListenAddress() string
}

type Parent interface {
	Error(error)
	OnNewConnection(net.Conn)
}

type Handler struct {
	parent   Parent
	config   Config
	log      *zap.Logger
	listener struct {
		sync.Mutex
		l net.Listener
	}
}

var (
	shutdown chan bool
)

func init() {
	shutdown = make(chan bool, 1)
}

func NewHandler(parent Parent, config Config, log *zap.Logger) Handler {
	return Handler{
		parent: parent,
		config: config,
		log:    log,
		listener: struct {
			sync.Mutex
			l net.Listener
		}{sync.Mutex{}, nil},
	}
}

func (h Handler) Start() {
	ln, err := net.Listen("tcp", h.config.ListenAddress())
	if err != nil {
		h.parent.Error(fmt.Errorf("failed to start listener: failed to start listening on address %s: %v", h.config.address, err))
	}

	h.listener.Lock()
	h.listener.l = ln
	h.listener.Unlock()

	h.log.Info("accept connections")
	for {
		select {
		case _ = <-shutdown:
			break

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
	h.log.Info("Listener stopped")
}

func (h Handler) Shutdown() {
	select {
	case shutdown <- true:
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
