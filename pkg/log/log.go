package log

import "github.com/sirupsen/logrus"

var (
	// std is the name of the standard logger in stdlib `log`
	std = logrus.New()
)

func Namespace(name string) *logrus.Logger {
	std.WithField("ns", name)
	return std
}
