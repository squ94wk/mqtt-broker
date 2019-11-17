package connect

var (
	actions chan func(Handler) error
)

type Parent interface {
	Error(error)
}

type Handler struct {
	parent   Parent
	shutdown chan bool
}

func (h Handler) Start() {
	for {
		select {
		case action := <-actions:
			err := action(h)
			if err != nil {
				h.parent.Error(err)
			}

		case _ = <-h.shutdown:
			break
		}
	}
}
