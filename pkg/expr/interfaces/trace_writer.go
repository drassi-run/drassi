package interfaces

type TraceWriter interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
}
