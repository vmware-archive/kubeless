package kubeless

import (
	"github.com/kubeless/kubeless/pkg/functions"
	"github.com/sirupsen/logrus"
)

// Hello sample function with dependencies
func Hello(event functions.Event, context functions.Context) (string, error) {
	logrus.Info(event.Data)
	return "Hello world!", nil
}
