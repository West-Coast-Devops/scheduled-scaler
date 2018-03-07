package cron

import (
	"time"
	"github.com/robfig/cron"
)

func Create(timeZone string) *cron.Cron {
	l, _ := time.LoadLocation(timeZone)
	
	return cron.NewWithLocation(l)
}

func Push(c *cron.Cron, time string, call func()) {
	s, _ := cron.Parse(time)
	c.Schedule(s, cron.FuncJob(call))
}

func Start(c *cron.Cron) {
	c.Start()
}
