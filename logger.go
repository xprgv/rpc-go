package rpc

type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

type fakeLogger struct{}

func (l fakeLogger) Debug(format string, args ...any) {}
func (l fakeLogger) Info(format string, args ...any)  {}
func (l fakeLogger) Warn(format string, args ...any)  {}
func (l fakeLogger) Error(format string, args ...any) {}
