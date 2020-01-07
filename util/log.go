package util

import (
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

// Add "file, line, function" context to the logger
func LogFLF(entry *logrus.Entry) *logrus.Entry {
	if pc, file, line, ok := runtime.Caller(1); ok {
		function := runtime.FuncForPC(pc).Name()
		return entry.WithFields(logrus.Fields{
			"file": path.Base(file),
			"line": line,
			"func": path.Base(function),
		})
	}
	return entry
}
