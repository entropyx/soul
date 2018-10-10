package engines

import (
	"testing"

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
