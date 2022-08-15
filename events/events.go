package events

type Event int

const (
	SaleCreated Event = iota
)

var handlers map[Event][]func(data interface{})

func Handle(event Event, handler func(data interface{})) {
	if handlers == nil {
		handlers = make(map[Event][]func(data interface{}))
	}
	handlers[event] = append(handlers[event], handler)
}

func Dispatch(event Event, data interface{}) {
	for _, eventHandlers := range handlers {
		for _, handler := range eventHandlers {
			handler(data)
		}
	}
}
