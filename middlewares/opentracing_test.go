package middlewares

import (
	"github.com/dsmontoya/soul/context"
	opentracing "github.com/opentracing/opentracing-go"

	"testing"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOpentracing(t *testing.T) {
	tr, _ := ConfigureDatadog("test")
	defer tracer.Stop()
	ConfigureOpenTracing(tr)

	defer Convey("Given a context", t, func() {
		c := context.NewContext(&context.MockContext{})

		Convey("When the handler is executed without a parent span", func() {
			Opentracing(c)

			Convey("The span should be set", func() {
				span := c.Get("span")
				So(span, ShouldNotBeNil)
			})

			Convey("The response header should contain a trace id", func() {
				So(c.Headers[datadogHeaderName], ShouldNotBeEmpty)
			})
		})

		Convey("When the handler is executed with a parent span", func() {
			span := tr.StartSpan("client")
			tr.Inject(span.Context(), opentracing.HTTPHeaders, c.Request.Headers)
			defer span.Finish()
			Opentracing(c)

			Convey("The span should be set", func() {
				span := c.Get("span")
				So(span, ShouldNotBeNil)
			})

			Convey("The request and response trace id headers should be the same", func() {
				So(c.Request.Headers[datadogHeaderName], ShouldEqual, c.Headers[datadogHeaderName])
			})
		})
	})
}
