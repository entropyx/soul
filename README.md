# Soul

Soul is an open source tool to easily create and deploy services in golang.

# Getting started

Let's get your service up and running in few minutes.

Required packages for this example:

    import (
    	"github.com/entropyx/soul"
    	"github.com/entropyx/soul/middlewares" //built-in middlewares
    	"github.com/entropyx/soulutils" //internal helper functions
    )

Initialize and name your new service.

    service := soul.New("myservice")

The default router uses AMQP as engine and includes important middlewares:

    r := soulutils.DefaultRouter(service)

# **Defining routes**

Create a group named after the main resource:

    users := r.Group("users")

And any subgroup you need:

    configuration := users.Group("configuration")

Generate a consumer that listens to the routing key and executes a function for every incoming message:

    users.Listen("update", func(c *context.Context) {
    	...
    })

Once the consumer is ready, call Service.Run. Everything below this function will run after the service is shutdown.

    service.Run()
    fmt.Println("Bye!")

Start the consumer by executing the `listen` command:

    myservice listen users.update

# Requests handling

A [context.Handler](https://godoc.org/github.com/entropyx/soul/context#Handler) receives a [context](https://godoc.org/github.com/entropyx/soul/context#Context) which contains relevant information about the current request and the required methods to handle it.

    func UpdateUsers(c *context.Context) {
    	...
    }

To properly get the request body use [Bind](https://godoc.org/github.com/entropyx/soul/context#Context.Bind).

    update := &protos.UpdateUsers{}
    if err := c.Bind(update); err != nil {
    	c.Error = err
    	return
    }

As you could notice any reached error should be set to c.Error, so it can be handled by pending middlewares. Don't forget to stop the execution of the current function in order to avoid unexpected behaviors.

If everything is okay, call `soulutils.WriteResponse` to write a proto to a `protos.Response`. This won't stop the current function.

    func UpdateUsers(c *context.Context) {
    	user := &protos.Users{}
    	soulutils.WriteResponse(c, user)
    }

For legacy response messages use [context.Proto](https://godoc.org/github.com/entropyx/soul/context#Context.Proto).

    c.Proto(user)

# Requesting other services

Services are intended to communicate across the network which can lead to unexpected behaviors. To troubleshoot services performance, propagate an Open Tracing context.

    func UpdateUsers(c *context.Context) {
    	usersService := services.NewUsers(services.OptionContext(c))
    }

# Logging

Get the [log entry](https://godoc.org/github.com/sirupsen/logrus#Entry) from [context.Log](https://godoc.org/github.com/entropyx/soul/context#Context.Log) to print a message:

    c.Log().Info("Hello")
    c.Log().Error("Something bad just happened")

You can assign the entry to a simpler variable:

    log := c.Log()
    log.Info("Hello")

The default entry includes all the fields that were previously configured by the middlewares. If you need to add some extra fields, pass the result of calling [entry.WithField](https://godoc.org/github.com/sirupsen/logrus#Entry.WithField) or [entry.WithFields](https://godoc.org/github.com/sirupsen/logrus#Entry.WithFields) over the default entry to [context.SetLog](https://godoc.org/github.com/entropyx/soul/context#Context.SetLog) . Passing a new entry will just rewrite the current one.

    // Use WithField if you need a single field.
    c.SetLog(c.Log().WithField("organization_id", organization.ID))
    // Use WithFields if you need more than a field.
    c.SetLog(c.Log().WithFields(logrus.Fields{
    	"user_id":user.ID,
    	"account_id":account.ID
    }))

## Getting the entry from standard contexts

There are some situation, like in commands and cronjobs, that will force you to work with a standard [context](https://godoc.org/context#Context).

Set a log entry to a context with `soulutils.WithLog`.

    import (
    	"context"
    	"github.com/entropyx/soulutils"
    	"github.com/sirupsen/logrus"
    )
    
    func handler(c context.Context) {
    	entry := logrus.NewEntry(logrus.New())
    	c = soulutils.WithLog(c,entry)
    }

This sets the value in the context using a custom key, which is only accessible through `soulutils.Log`.

    log := soulutils.Log(c)
    log.Info("Hello")

# Health check

By default, your consumers will report their own status. If your consumer is successfully connected, the http status code will be 200 and the body will include a health check status list:

Your service can report the status of database connections or external services availability.

    curl localhost:8081/health_check

    HTTP/1.1 200 OK

    {"test.a 0":true}

You can add custom health checks:

    service.HealthCheck("database", func() bool {
    	if err := db.Ping(); err != nil {
    		return false
    	}
    	return true
    })

    {"database":true,"test.a 0":true}

If any of your health checks fails, the status code will be 500.

    HTTP/1.1 500 Internal Server Error

# Graceful shutdown

You can safely shutdown your service using quit command. Your consumer will stop receiving messages and disconnect itself once the current work is done.

    myservice quit

# Testing

Testing your handlers is a unavoidable task you must perform.

Required packages for the examples:

    import (
    	"github.com/entropyx/anyutils"
    	"github.com/entropyx/protos"
    	"github.com/entropyx/soul/context"
    	"github.com/entropyx/soulutils"
    	. "github.com/smartystreets/goconvey/convey"
    )

`soulutils.TestContext` simulates a request by sending `message` through the `default middlewares` and `handlers`.

    request := &protos.Users{Email: "daniel@entropy.tech"}
    c, mock := soulutils.TestContext(request, UpdateUsers)

The returned values are two:

**c *[context.Context](https://godoc.org/github.com/entropyx/soul/context#Context)**: This is the context passed through the handlers, which may contain an `error` (c.Error) and values ([c.Get](https://godoc.org/github.com/entropyx/soul/context#Context.Get)).

**mock *[context.MockContext](https://godoc.org/github.com/entropyx/soul/context#MockContext)**: The mocked engine context, which contains a response (*[context.R](https://godoc.org/github.com/entropyx/soul/context#R)).

Use `anyutils.UnmarshalResponse` to get the response body from `mock`:

    response := &protos.Users{}
    err := anyutils.UnmarshalResponse(mock.Response.Body, response)

Now you can test that your handler works as expected:

    Convey("The context error should be nil", func() {
    	So(c.Error, ShouldBeNil)
    })
    
    Convey("The response should be valid", func() {
    	So(response.Users[0].Email, ShouldEqual, "daniel@entropy.tech")
    })

# Commands

Your service is powered by [cobra](https://github.com/spf13/cobra), a tool for creating CLI applications. You can add new commands with [Service.Command](https://godoc.org/github.com/entropyx/soul#Service.Command). This implementation is easy and flexible enough but we need to add some custom code to each of the new commands. Use `soulutils.Command` to add extra configuration to a [context](https://godoc.org/context#Context).+

    import (
    	"context"
    	"github.com/entropyx/soul"
    	"github.com/spf13/cobra"
    )
    
    var cmdListUsers = &cobra.Command{
    	Use: "list",
    	Short: "List the available users",
    	Run: soulutils.Command(ListUsers),
    }
    
    func main(){
    	service := soul.New("moises")
    	service.Command(cmdListUsers)
    }

It generates a `func(*cobra.Command, []string)` that is used as a [cobra.Command](https://godoc.org/github.com/spf13/cobra#Command) run function. You can get the function parameters with `soulutils.CmdValues`:

    func ListUsers(c context.Context){
    	cmd,args := soulutils.CmdValues(c)
    	...
    }

## Logging

    log := soulutils.Log(c)
    log.Info("Hello")

## Services

    spanCtx := soulutils.SpanContext(c)
    service := services.New(services.OptionSpanContext(spanCtx))
    organizationsService := services.NewOrganizations(service)

## Warning

`soulutils.Command` API will probably change, since a hook implementation may be the best idiomatic alternative. Stay tunned, please.

# Cron jobs

A cron job is a task that runs periodically at fixed intervals. You can schedule your own cronjob with [service.Cronjob](https://godoc.org/github.com/entropyx/soul#Service.CronJob). Although, like in the commands, you need to add some custom code.

Your final cron job function must receive a context used to share some values. You will see how to get those values in short.

    func UpdateUsers(c context.Context){
    	...
    }

Now, pass the name of your cron job and your new function wrapped with `soulutils.CronJob` to [service.Cronjob](https://godoc.org/github.com/entropyx/soul#Service.CronJob).

    service.CronJob("update-users", soulutils.CronJob(UpdateUsers))

Run your cron job with `cronjob` command. Define the interval with `-s`. The schedule format is explained in [cron](https://godoc.org/github.com/robfig/cron) documentation.

    myservice cronjob -s @hourly update-users

## Logging

    log := soulutils.Log(c)
    log.Info("Hello")

## Services

    spanCtx := soulutils.SpanContext(c)
    service := services.New(services.OptionSpanContext(spanCtx))
    organizationsService := services.NewOrganizations(service)

# Add your service to services

Your service struct should embed a pointer to a Service, which includes methods to publish.

    type Users struct {
    	*Service
    }

Create a function that returns your preconfigured service struct. This function should first initialize a Service by calling New and set it to your struct.

    func NewUsers(options ...Option) *Users {
    	s := New(options...)
    	return &Users{s}
    }

An Option is a function that receives the \*Service you initialize. It modifies the Service state.

    services.NewUsers(services.OptionSpanContext(spanCtx))

Create your action method. unmarshal returns the response error.

    func (u *Users) Retrieve(selector *protos.Selector) ([]*protos.User,error) {
    	var body []byte
    	users := &protos.Users{}
    	err := u.publish(&publishingConfig{
    		routingKey: "users.retrieve",
    		timeout: u.Timeout,
    		out: selector
    	}, func(delivery *rabbitgo.Delivery) {
    			body = delivery.Body
    	})
    	
    	if err != nil {
    		return nil, err
    	}
    	
    	if err := unmarshal(body, users); err != nil {
    		return nil, err
    	}
    	return users.Items, nil
    }
