package ast_ifaces

type TraceWriter interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
}
