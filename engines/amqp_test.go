package engines

import (
	"testing"

	"github.com/entropyx/rabbitgo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMergeRoutingKeys(t *testing.T) {
	Convey("Given an amqp engine", t, func() {
		amqp := &AMQP{}

		Convey("When two routing keys are merged", func() {
			routingKey := "logs.warning"
			r := amqp.MergeRoutingKeys(routingKey, "#")

			Convey("The final routing key should be 'logs.warning.#'", func() {
				So(r, ShouldEqual, "logs.warning.#")
			})
		})

	})
}

func TestIsConnected(t *testing.T) {
	Convey("Given an amqp engine with a connection", t, func() {
		conn := &rabbitgo.Connection{}
		amqp := &AMQP{conn: conn}

		Convey("When the connection is active and unblocked", func() {
			conn.IsConnected = true
			Convey("The engine should NOT be connected", func() {
				So(amqp.IsConnected(), ShouldBeTrue)
			})
		})

		Convey("When the connection is active and blocked", func() {
			conn.IsConnected = true
			conn.IsBlocked = true

			Convey("The engine should NOT be connected", func() {
				So(amqp.IsConnected(), ShouldBeFalse)
			})
		})

		Convey("When the connection is unactive and blocked", func() {
			conn.IsBlocked = true

			Convey("The engine should NOT be connected", func() {
				So(amqp.IsConnected(), ShouldBeFalse)
			})
		})

		Convey("When the connection is unactive and unblocked", func() {
			Convey("The engine should NOT be connected", func() {
				So(amqp.IsConnected(), ShouldBeFalse)
			})
		})
	})
}
