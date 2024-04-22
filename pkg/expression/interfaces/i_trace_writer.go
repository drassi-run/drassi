package interfaces

type ITraceWriter interface {
	Info(msg string)
	Verbose(msg string)
}
