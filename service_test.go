package soul

import (
	"testing"
	"time"

	"github.com/dsmontoya/soul/context"
	"github.com/spf13/cobra"

	"github.com/dsmontoya/soul/engines"
	. "github.com/smartystreets/goconvey/convey"
)

func TestListen(t *testing.T) {
	Convey("Given a service with routes", t, func() {
		mock := &engines.Mock{}
		service := New("test")
		router := service.NewRouter(mock)
		router.Include(func(c *context.Context) {})
		router.Listen("logs.warning", func(c *context.Context) {})

		Convey("When the routes are listened", func() {
			service.listen(&cobra.Command{}, []string{"logs.warning"})
			time.Sleep(1 * time.Millisecond)

			Convey("The engine should be connected", func() {
				So(mock.IsConnected, ShouldBeTrue)
			})

			Convey("The listened routing key should be 'logs.warning'", func() {
				So(mock.RoutingKey, ShouldEqual, "logs.warning")
			})

			Convey("The number of handlers should be 2", func() {
				So(mock.Handlers, ShouldHaveLength, 2)
			})
		})
	})
}
