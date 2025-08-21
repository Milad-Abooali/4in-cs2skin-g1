package events

type Event struct {
	Target string
	UserID int64
	Type   string
	Data   interface{}
}

var Bus = make(chan Event, 100)

func Emit(target string, eventType string, data interface{}) {
	ev := Event{
		Target: target,
		Type:   eventType,
		Data:   data,
	}
	select {
	case Bus <- ev:
	default:

	}
}
