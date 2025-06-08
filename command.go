package workerpool

type actionType int

const (
	addWorker actionType = iota
	removeWorker
)

type command struct {
	action actionType
	done   chan error
}

func newCommand(action actionType) command {
	return command{
		action: action,
		done:   make(chan error),
	}
}
