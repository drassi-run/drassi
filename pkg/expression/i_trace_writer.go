package expression

type ITraceWriter interface {
	Info(msg string)
	Verbose(msg string)
}
