package tracers

import (
	"testing"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

type testWriter struct {
	data []byte
}

func (t *testWriter) Write(p []byte) (int, error) {
	t.data = p
	return len(p), nil
}

func TestStackdriverFormatter(t *testing.T) {
	Convey("Given a stackdriver formatter with a json format", t, func() {
		formatter := &StackdriverFormatter{}
		Convey("When you write a log", func() {
			logger := logrus.New()
			output := &testWriter{}
			logger.SetFormatter(formatter)
			logger.SetOutput(output)
			entry := logrus.NewEntry(logger)
			entry.Info("hello")

			Convey("The output should include the severity field", func() {
				So(string(output.data), ShouldContainSubstring, "\"severity\":\"info\"")
			})
		})
	})
}
