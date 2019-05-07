package soul

import (
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

type cronJob struct {
	name    string
	handler func()
}

func (c *cronJob) Start(spec string) {
	log.Info("Scheduling cronjob " + spec)
	cj := cron.New()
	cj.AddFunc(spec, func() {
		log.Info("Running cronjob " + c.name)
		c.handler()
	})
	cj.Start()
}
