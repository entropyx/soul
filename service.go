package soul

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/engines"
	"github.com/entropyx/soul/env"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var schedule string
var vars []string

const (
	cmdLongCronJob  = "Run task on the given schedule"
	cmdShortCronJob = "Run task on the given schedule"

	cmdShortListener = "Listen to messages in a routing key"
	cmdLongListener  = "Bind a consumer to a routing key. Every time a new message is received the consumer will execute the attached handlers."

	cmdShortListenerList = "List the available routing keys"

	cmdLongQuit  = "Gracefully shutdown"
	cmdShortQuit = "Gracefully shutdown"
)

type HealthCheck func() bool

type Service struct {
	healthChecks map[string]HealthCheck
	rootCmd      *cobra.Command
	routers      []*Router
	cronJobs     map[string]*cronJob
	commands     []*cobra.Command
	consumers    []engines.Consumer
	close        chan uint8
	flags        flags
}

type flags struct {
	routingKey string
}

func New(name string) *Service {
	env.Name = name
	rootCmd := &cobra.Command{
		Use:   env.Name,
		Short: "",
		Long:  "",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	rootCmd.PersistentFlags().StringArrayVarP(&vars, "variables", "v", []string{}, "set variable list")

	return &Service{healthChecks: map[string]HealthCheck{}, rootCmd: rootCmd, cronJobs: map[string]*cronJob{}, close: make(chan uint8)}
}

func (s *Service) Command(command *cobra.Command) {
	s.commands = append(s.commands, command)
}

func (s *Service) CronJob(name string, handler func()) {
	s.cronJobs[name] = &cronJob{name, handler}
}

func (s *Service) HealthCheck(name string, healthCheck HealthCheck) {
	s.healthChecks[name] = healthCheck
}

func (s *Service) NewRouter(engine engines.Engine) *Router {
	router := &Router{engine: engine, routes: map[string][]context.Handler{}}
	router.router = router
	s.routers = append(s.routers, router)
	return router
}

func (s *Service) Run() {
	s.addCommandCronJob()
	s.addCommandListener()
	s.addCommandQuit()
	s.addCommands()
	s.healthCheckConsumer()
	if err := s.rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func (s *Service) addCommandCronJob() {
	cronJobCmd := &cobra.Command{
		Use:   "cronjob job",
		Short: cmdShortCronJob,
		Args:  cobra.ExactArgs(1),
		Run:   s.cronjob,
	}
	cronJobCmd.Flags().StringVarP(&schedule, "schedule", "s", "1h", "Run cron job task on the time")
	s.rootCmd.AddCommand(cronJobCmd)
}

func (s *Service) addCommandListener() {
	listenCmd := &cobra.Command{
		Use:   "listen routing_key",
		Short: cmdShortListener,
		Args:  cobra.ExactArgs(1),
		Run:   s.listen,
	}
	listenListCmd := &cobra.Command{
		Use:   "list",
		Short: cmdShortListenerList,
		Run:   s.listenList,
	}
	listenCmd.AddCommand(listenListCmd)
	s.rootCmd.AddCommand(listenCmd)
}

func (s *Service) addCommandQuit() {
	cronJobCmd := &cobra.Command{
		Use:   "quit",
		Short: cmdShortQuit,
		Run:   quit,
	}
	cronJobCmd.Flags().StringVarP(&schedule, "schedule", "s", "1h", "Run cron job task on the time")
	s.rootCmd.AddCommand(cronJobCmd)
}

func (s *Service) addCommands() {
	s.rootCmd.AddCommand(s.commands...)
}

func (s *Service) cronjob(cmd *cobra.Command, args []string) {
	job := args[0]
	forever := make(chan bool)
	if cronjob, ok := s.cronJobs[job]; ok {
		cronjob.Start(schedule)
		<-forever
	}
	log.Fatal("Invalid job")
}

func (s *Service) gracefulShutdownConsumer() {
	engine := &engines.HTTPSimple{Address: ":8081"}
	engine.Connect()
	gracefulShutdown, _ := engine.Consumer("/graceful_shutdown")
	gfsHandlers := []context.Handler{
		func(c *context.Context) {
			c.Log().Info("Graceful shutdown")
			s.close <- 0
			c.Headers["status"] = 201
			c.String("done")
		},
	}
	gracefulShutdown.Consume(gfsHandlers)
	s.consumers = append(s.consumers, gracefulShutdown)
	log.Info("Graceful shutdown configured")
}

func (s *Service) healthCheckConsumer() {
	engine := &engines.HTTPSimple{Address: ":8081"}
	engine.Connect()
	healthCheck, _ := engine.Consumer("/health_check")
	hckHandlers := []context.Handler{
		func(c *context.Context) {
			c.Log().Debug("Health check")
			pass := true
			m := map[string]bool{}
			for k, f := range s.healthChecks {
				v := f()
				if !v {
					pass = false
				}
				m[k] = v
			}
			defer c.JSON(m)
			if !pass {
				c.Headers["status"] = 500
				return
			}
			c.Headers["status"] = 200
		},
	}
	healthCheck.Consume(hckHandlers)
	s.consumers = append(s.consumers, healthCheck)
	log.Info("Health check configured")
}

func (s *Service) listenAll() {

}

func (s *Service) listen(cmd *cobra.Command, args []string) {
	routingKey := args[0]
	s.listenRouters(routingKey)
	log.Printf("Waiting for messages. To exit press CTRL+C")
	s.notifyInterrupt()
	s.gracefulShutdownConsumer()
	code := <-s.close
	log.Infof("%s is shutting down", env.Name)
	s.shutdown()
	log.Println("Exit with code", code)
}

func (s *Service) notifyInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		sig := <-c
		log.Info("Signal ", sig)
		s.close <- 0
	}()
}

func (s *Service) listenList(cmd *cobra.Command, args []string) {
	for routingKey, handlers := range s.routes() {
		fmt.Printf("%s: %d handlers\n", routingKey, len(handlers))
	}
}

func (s *Service) listenRouters(routingKey string) {
	m := map[string]uint8{}
	for _, router := range s.routers {
		if handlers, ok := router.routes[routingKey]; ok {
			v := m[routingKey]
			name := fmt.Sprintf("%s %d", routingKey, v)
			m[routingKey]++
			routerCopy := *router
			s.HealthCheck(name, HealthCheck(func() bool {
				return routerCopy.engine.IsConnected()
			}))
			go func() {
				if !routerCopy.engine.IsConnected() {
					routerCopy.connect()
				}
				consumer, err := routerCopy.engine.Consumer(routingKey)
				if err != nil {
					panic(err)
				}
				s.consumers = append(s.consumers, consumer)
				log.Info("Starting handler")
				if err := consumer.Consume(handlers); err != nil {
					log.Error("Unable to consume", err)
				}
			}()
			continue
		}
		log.Error("Handler not found")
	}
}

func (s *Service) routes() map[string][]context.Handler {
	routes := map[string][]context.Handler{}
	for _, router := range s.routers {
		for routingKey, handlers := range router.routes {
			routes[routingKey] = append(routes[routingKey], handlers[len(handlers)-1])
		}
	}
	return routes
}

func (s *Service) shutdown() {
	for _, consumer := range s.consumers {
		err := consumer.Close()
		if err != nil {
			log.Errorf("Error while closing consumer: %s", err.Error())
			continue
		}
	}
	for _, router := range s.routers {
		err := router.engine.Close()
		if err != nil {
			log.Errorf("Error while closing connection: %s", err.Error())
			continue
		}
	}
}

func quit(cmd *cobra.Command, args []string) {
	_, err := http.Get("http://localhost:8081/graceful_shutdown")
	if err != nil {
		log.Fatal("Unable to shutdown. Is the service running?")
		return
	}
	log.Info("Shutting down")
}
