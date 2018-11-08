package soul

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/entropyx/soul/context"
	"github.com/spf13/cobra"
)

const (
	cmdShortListener = "Listen to messages in a routing key"
	cmdLongListener  = "Bind a consumer to a routing key. Every time a new message is received the consumer will execute the attached handlers."

	cmdShortListenerList = "List the available routing keys"
)

type Service struct {
	Name     string
	rootCmd  *cobra.Command
	routers  []*Router
	cronJobs map[string]*cronJob
	commands []*cobra.Command
	flags    flags
}

type flags struct {
	routingKey string
}

func New(name string) *Service {
	rootCmd := &cobra.Command{
		Use:   name,
		Short: "",
		Long:  "",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	return &Service{Name: name, rootCmd: rootCmd, cronJobs: map[string]*cronJob{}}
}

func (s *Service) Command(command *cobra.Command) {
	s.commands = append(s.commands, command)
}

func (s *Service) CronJob(name, spec string, handler func()) {
	s.cronJobs[name] = &cronJob{name, spec, handler}
}

func (s *Service) NewRouter(engine Engine) *Router {
	router := &Router{engine: engine, routes: map[string][]context.Handler{}}
	router.router = router
	s.routers = append(s.routers, router)
	return router
}

func (s *Service) Run() {
	forever := make(chan bool)
	s.addCommandListener()
	s.addCommands()
	if err := s.rootCmd.Execute(); err != nil {
		panic(err)
	}
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func (s *Service) addCommandCronJob() {
	cronJobCmd := &cobra.Command{
		Use:   "cronjob job",
		Short: cmdShortListener,
		Args:  cobra.ExactArgs(1),
		Run:   s.cronjob,
	}
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

func (s *Service) addCommands() {
	s.rootCmd.AddCommand(s.commands...)
}

func (s *Service) cronjob(cmd *cobra.Command, args []string) {
	job := args[0]
	if cronjob, ok := s.cronJobs[job]; ok {
		cronjob.Start()
		return
	}
	log.Fatal("Invalid job")

}

func (s *Service) listenAll() {

}

func (s *Service) listen(cmd *cobra.Command, args []string) {
	routingKey := args[0]
	for _, router := range s.routers {
		router.connect()
		if handlers, ok := router.routes[routingKey]; ok {
			log.Info("Starting handler")
			go router.engine.Consume(routingKey, handlers)
		}
	}

}

func (s *Service) listenList(cmd *cobra.Command, args []string) {
	for routingKey, handlers := range s.routes() {
		fmt.Printf("%s: %d handlers\n", routingKey, len(handlers))
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
