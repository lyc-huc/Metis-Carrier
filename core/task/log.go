package task

import "github.com/sirupsen/logrus"

// Global log object, used by the current package.
var log = logrus.WithField("prefix", "task")

