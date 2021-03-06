package logger

type Logger interface {
	Info(msg string)
	Debug(msg string)
	Error(err error, msg string)
}
