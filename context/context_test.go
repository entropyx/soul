package context

import (
	"math"
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func TestIndex(t *testing.T) {
	c.Convey("Given a list of handlers", t, func() {
		var index uint8
		f := func(c *Context) {
			index++
		}
		handlers := []Handler{f, f, func(c *Context) { c.Abort("") }, f}

		c.Convey("When the handlers run", func() {
			context := &Context{Request: &R{}}
			context.RunHandlers(handlers...)

			c.Convey("The index should be 2", func() {
				c.So(index, c.ShouldEqual, 2)
			})

			c.Convey("The context index should be the abort index", func() {
				c.So(context.index, c.ShouldEqual, math.MaxInt8)
			})
		})
	})
}
