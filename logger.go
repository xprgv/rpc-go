package rpc

type Logger interface {
	Trace(args ...interface{})
	Tracef(format string, args ...interface{})

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}

type fakeLogger struct{}

func (l fakeLogger) Trace(args ...interface{})                 {}
func (l fakeLogger) Tracef(format string, args ...interface{}) {}

func (l fakeLogger) Debug(args ...interface{})                 {}
func (l fakeLogger) Debugf(format string, args ...interface{}) {}

func (l fakeLogger) Info(args ...interface{})                 {}
func (l fakeLogger) Infof(format string, args ...interface{}) {}

func (l fakeLogger) Warn(args ...interface{})                 {}
func (l fakeLogger) Warnf(format string, args ...interface{}) {}

func (l fakeLogger) Error(args ...interface{})                 {}
func (l fakeLogger) Errorf(format string, args ...interface{}) {}

func (l fakeLogger) Fatal(args ...interface{})                 {}
func (l fakeLogger) Fatalf(format string, args ...interface{}) {}
