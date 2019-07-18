package log

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test(t *testing.T) {
	Convey("Given the severity info", t, func() {
		s := &Stackdriver{severity: Info}
		Convey("When canLog is called with a debug severity", func() {
			ok := s.canLog(Debug)

			Convey("The returned value should be false", func() {
				So(ok, ShouldBeFalse)
			})
		})

		Convey("When canLog is called with a warning severity", func() {
			ok := s.canLog(Warning)

			Convey("The returned value should be false", func() {
				So(ok, ShouldBeTrue)
			})
		})
	})
}
