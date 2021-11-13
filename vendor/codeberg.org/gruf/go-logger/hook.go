package logger

// Hook defines a log Entry modifier
type Hook interface {
	Do(*Entry)
}

// HookFunc is a simple adapter to allow functions to satisfy the Hook interface
type HookFunc func(*Entry)

func (hook HookFunc) Do(entry *Entry) {
	hook(entry)
}
