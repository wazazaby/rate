package rate

import "fmt"

type Logger interface {
	Infof(format string, args ...any)
	Debugf(format string, args ...any)
	Errorf(format string, args ...any)
}

// discardLogger is [Logger] implementation that does nothing with it's input.
type discardLogger struct{}

func (discardLogger) Infof(string, ...any)  {}
func (discardLogger) Debugf(string, ...any) {}
func (discardLogger) Errorf(string, ...any) {}

// fmtLogger is a [Logger] implementation that prints to [os.Stdout] using a
// custom formatting template, passed to [fmt.Printf].
type fmtLogger struct{}

func (fmtLogger) Infof(format string, args ...any) {
	fmt.Printf("[INFO]  "+format+"\n", args...)
}
func (fmtLogger) Debugf(format string, args ...any) {
	fmt.Printf("[DEBUG] "+format+"\n", args...)
}
func (fmtLogger) Errorf(format string, args ...any) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

var (
	noopLogger     = &discardLogger{}
	standardLogger = &fmtLogger{}
)

var (
	_ Logger = &discardLogger{}
	_ Logger = &fmtLogger{}
)
