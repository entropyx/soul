package soul

import (
	"testing"

	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/engines"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetValue(t *testing.T) {
	Convey("Given a list of variables", t, func() {
		vars = append(vars, "1=a", "2=b")

		Convey("When the value 2 is got", func() {
			v := GetValue("2")

			Convey("The value should be 'b'", func() {
				So(v, ShouldEqual, "b")
			})
		})
	})
}

func TestInclude(t *testing.T) {
	Convey("Given a group", t, func() {
		service := &Service{}
		router := service.NewRouter(&engines.AMQP{})
		logs := router.Group("logs")

		Convey("When a handler is included", func() {
			logs.Include(func(c *context.Context) {})

			Convey("The number of handlers should be 1", func() {
				So(logs.handlers, ShouldHaveLength, 1)
			})

			Convey("When a handler is included to a subgroup", func() {
				warning := logs.Group("warning")
				warning.Include(func(c *context.Context) {})

				Convey("The number of handlers should be 2", func() {
					So(warning.handlers, ShouldHaveLength, 2)
				})
			})
		})

	})
}

func TesRouterGrouptListen(t *testing.T) {
	Convey("Given a group", t, func() {
		service := &Service{}
		router := service.NewRouter(&engines.AMQP{})
		logs := router.Group("logs")

		Convey("When a routing key is listened", func() {
			logs.Listen("warning", func(c *context.Context) {

			})

			Convey("The number of service routes should be 1", func() {
				So(router.routes, ShouldHaveLength, 1)
			})

			Convey("The handler exist", func() {
				So(router.routes["logs.warning"], ShouldNotBeNil)
			})
		})
	})
}

func TestRouter(t *testing.T) {
	Convey("Given a log service", t, func() {
		service := &Service{}
		Convey("When a router is initialized", func() {
			router := service.NewRouter(&engines.AMQP{})

			Convey("Given a 'logs' routing key", func() {
				logs := router.Group("logs")

				Convey("The routing key should be 'logs'", func() {
					So(logs.routingKey, ShouldEqual, "logs")
				})

				Convey("When a handler is included", func() {
					logs.Include(func(c *context.Context) {})

					Convey("The handlers list should contains one handler", func() {
						So(logs.handlers, ShouldHaveLength, 1)
					})
				})

				Convey("When a 'logs.warning' routing key is generated", func() {
					warning := logs.Group("warning")

					Convey("The routing key should be logs.error", func() {
						So(warning.routingKey, ShouldEqual, "logs.warning")
					})
				})
			})
		})
	})
}

func Test_mergeRoutingKey(t *testing.T) {
	Convey("Given a route group", t, func() {
		router := &Router{engine: &engines.AMQP{}}
		group := &RouteGroup{
			router:     router,
			routingKey: "logs.warning",
		}

		Convey("When the routing key is merged", func() {
			r := group.mergeRoutingKey("#")

			Convey("The final routing key should be 'logs.warning.#'", func() {
				So(r, ShouldEqual, "logs.warning.#")
			})
		})
	})
}
