package actor

import (
	"fmt"
	"net"
)

type ConnectionParent interface {
	Error(error)
	NewConnection(net.Conn)
}

type Listener struct {
	address string
	parent  ConnectionParent
}

//func NewListener(address string, parent Parent) *Listener {
//return &Listener{
//address: address,
//parent:  parent,
//}
//}

func (l *Listener) Start() {
	ln, err := net.Listen("tcp", l.address)
	if err != nil {
		l.parent.Error(fmt.Errorf("failed to start listener: failed to start listening on port %s: %v", l.port, err))
	}

	for {
		log.Infof("accept connections")
		conn, err := ln.Accept()
		if err != nil {
			l.parent.Error(fmt.Errorf("failed to accept connections: %v", err))
			break
		}

		log.Infof("new connection")
		l.parent.NewConnection(conn)
	}

	log.Info("Listener stopped")
}

func (l *Listener) Close() error {
	err := l.Close()
	if err != nil {
		return fmt.Errorf("failed to close Listener: %v", err)
	}

	return nil
}
