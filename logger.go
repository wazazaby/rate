package rate

import "fmt"

type Logger interface {
	Infof(format string, args ...any)
	Debugf(format string, args ...any)
	Errorf(format string, args ...any)
}

type noopLoggerImpl struct{}

func (*noopLoggerImpl) Infof(string, ...any)  {}
func (*noopLoggerImpl) Debugf(string, ...any) {}
func (*noopLoggerImpl) Errorf(string, ...any) {}

type standardLoggerImpl struct{}

func (*standardLoggerImpl) Infof(format string, args ...any) {
	fmt.Printf("[INFO]  "+format+"\n", args...)
}
func (*standardLoggerImpl) Debugf(format string, args ...any) {
	fmt.Printf("[DEBUG] "+format+"\n", args...)
}
func (*standardLoggerImpl) Errorf(format string, args ...any) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

var (
	noopLogger     = &noopLoggerImpl{}
	standardLogger = &standardLoggerImpl{}
)
