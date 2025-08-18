package monitor

type ConnectionStatus int

const (
	Running ConnectionStatus = iota
	Slow
	Down
	Inactive
)

func (c ConnectionStatus) String() string {
	switch c {
	case Running:
		return "RUNNING"
	case Slow:
		return "SLOW"
	case Down:
		return "DOWN"
	case Inactive:
		return "INACTIVE"
	default:
		return "UNKNOWN"
	}
}
