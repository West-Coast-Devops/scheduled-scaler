package cron

import (
	"github.com/robfig/cron/v3"
	"time"
)

var (
	V1Parser = cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
)

func Parse(spec string) (cron.Schedule, error) {
	return V1Parser.Parse(spec)
}

func Create(timeZone string) *cron.Cron {
	l, _ := time.LoadLocation(timeZone)

	return cron.New(cron.WithLocation(l), cron.WithParser(V1Parser))
}

func Push(c *cron.Cron, time string, call func()) (cron.EntryID, error) {
	return c.AddFunc(time, call)
}

func Start(c *cron.Cron) {
	c.Start()
}

func Stop(c *cron.Cron) {
	c.Stop()
}
