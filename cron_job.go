package soul

import (
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

type cronJob struct {
	name    string
	spec    string
	handler func()
}

func (c *cronJob) Start() {
	log.Info("Starting cronjob")
	cj := cron.New()
	cj.AddFunc(c.spec, func() {
		log.Info("Running cronjob " + c.name)
		c.handler()
	})
	cj.Start()
}
