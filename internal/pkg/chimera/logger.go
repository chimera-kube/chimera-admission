package chimera

import (
	"flag"

	"k8s.io/klog/v2"
)

type Logger struct{}

func NewLogger(debug bool) Logger {
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	if debug {
		klogFlags.Set("v", "2")
	}

	return Logger{}
}

func (*Logger) Debug(msg string) {
	klog.V(2).Infoln(msg)
}

func (*Logger) Debugf(format string, args ...interface{}) {
	klog.V(2).Infof(format, args...)
}

func (*Logger) Info(msg string) {
	klog.Info(msg)
}

func (*Logger) Infof(format string, args ...interface{}) {
	klog.Infof(format, args...)
}
func (*Logger) Error(msg string) {
	klog.Error(msg)
}

func (*Logger) Errorf(format string, args ...interface{}) {
	klog.Errorf(format, args...)
}
