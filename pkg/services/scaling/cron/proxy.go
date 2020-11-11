package cron

//go:generate mockgen -source $GOFILE -destination=mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE

import (
	"github.com/robfig/cron"
	"time"
)

type CronProxy interface {
	Create(timeZone string) *cron.Cron
	Push(c *cron.Cron, time string, call func())
	Start(c *cron.Cron)
	Stop(c *cron.Cron)
}

type CronImpl struct {
}

func (ci *CronImpl) Create(timeZone string) *cron.Cron {
	l, _ := time.LoadLocation(timeZone)

	return cron.NewWithLocation(l)
}

func (ci *CronImpl) Push(c *cron.Cron, time string, call func()) {
	s, _ := cron.Parse(time)
	c.Schedule(s, cron.FuncJob(call))
}

func (ci *CronImpl) Start(c *cron.Cron) {
	c.Start()
}

func (ci *CronImpl) Stop(c *cron.Cron) {
	c.Stop()
}
