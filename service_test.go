package soul

import (
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/entropyx/soul/context"
	"github.com/spf13/cobra"

	"github.com/entropyx/soul/engines"
	. "github.com/smartystreets/goconvey/convey"
)

type cronJobMock struct {
	wg     *sync.WaitGroup
	called bool
}

func (c *cronJobMock) Run() {
	c.called = true
	c.wg.Done()
}

func TestListen(t *testing.T) {
	Convey("Given a service with routes", t, func() {
		mock := &engines.Mock{}
		service := New("test")
		router := service.NewRouter(mock)
		router.Include(func(c *context.Context) {})
		router.Listen("logs.warning", func(c *context.Context) {})

		Convey("When the routes are listened", func() {
			service.listenRouters("logs.warning")
			time.Sleep(10 * time.Millisecond)
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

		Convey("When the listen command is executed an a close signal is sent", func() {
			go service.listen(&cobra.Command{}, []string{"logs.warning"})
			time.Sleep(1 * time.Millisecond)
			service.close <- 1
			time.Sleep(1 * time.Millisecond)

			Convey("The engine should be closed", func() {
				So(mock.IsConnected, ShouldBeFalse)
			})

			Convey("The consumers should be closed", func() {
				for _, c := range service.consumers {
					consumer, ok := c.(*engines.MockConsumer)
					if ok {
						So(consumer.IsConnected, ShouldBeFalse)
					}
				}
			})
		})
	})
}

func Test_notifyInterrupt(t *testing.T) {
	Convey("Given a sigint listener", t, func() {
		service := &Service{close: make(chan uint8)}
		service.notifyInterrupt()

		Convey("When a interrupt signal is sent", func() {
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			c := <-service.close

			Convey("A close signal should be sent", func() {
				So(c, ShouldEqual, 0)
			})
		})

	})
}

func TestCronJob(t *testing.T) {
	Convey("Given a service with a cronjob", t, func() {
		wg := &sync.WaitGroup{}
		wg.Add(2)
		mock := &cronJobMock{wg: wg}
		service := New("test")
		schedule = "@every 2ms"
		service.CronJob("test", mock.Run)

		Convey("When the cronjob starts", func() {
			go service.cronjob(&cobra.Command{}, []string{"test"})
			wg.Done()
			wg.Wait()

			Convey("The cronjob should be called", func() {
				So(mock.called, ShouldBeTrue)
			})
		})
	})
}
