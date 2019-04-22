package engines

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/entropyx/soul/context"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConsume(t *testing.T) {
	Convey("Given a http engine consumer", t, func() {
		handlers := []context.Handler{
			func(c *context.Context) {
				c.Headers["status"] = 200
				c.Headers["test"] = "a"
				c.String("response")
			},
		}
		engine := &HTTPSimple{Address: ":8081"}
		engine.Connect()
		consumer, _ := engine.Consumer("/resource")

		Convey("When it is consumed", func() {
			go consumer.Consume(handlers)
			time.Sleep(10 * time.Millisecond)

			Convey("And a valid request is sent", func() {
				resp, err := http.Get("http://localhost:8081/resource")
				So(err, ShouldBeNil)
				Convey("The response should be valid", func() {
					So(resp.StatusCode, ShouldEqual, 200)
					So(resp.Header["Test"], ShouldResemble, []string{"a"})
					body, _ := ioutil.ReadAll(resp.Body)
					So(string(body), ShouldEqual, "response")
				})
			})

			Convey("And an invalid request is sent", func() {
				resp, _ := http.Get("http://localhost:8081/other_resource")
				Convey("The status code should be 404", func() {
					So(resp.StatusCode, ShouldEqual, 404)
				})
			})
		})
	})
}

func TestMergingRoutingKeys(t *testing.T) {
	Convey("Given an empty absolute routing key", t, func() {
		absolute := ""
		engine := &HTTPSimple{}

		Convey("When a relative key is merged", func() {
			merge := engine.MergeRoutingKeys(absolute, "resource")

			Convey("The final routing key should be '/resource'", func() {
				So(merge, ShouldEqual, "/resource")
			})
		})
	})

	Convey("Given a non empty absolute routing key", t, func() {
		absolute := "/resource/:id/other_resource"
		engine := &HTTPSimple{}

		Convey("When a relative key is merged", func() {
			merge := engine.MergeRoutingKeys(absolute, ":id")

			Convey("The final routing key should be valid", func() {
				So(merge, ShouldEqual, absolute+"/:id")
			})
		})
	})
}
