package cron

//go:generate mockgen -source=$GOFILE -destination=mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE

import (
	"github.com/robfig/cron/v3"
	"time"
)

var (
	V1Parser = cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
)

type CronProxy interface {
	Parse(spec string) (cron.Schedule, error)
	Create(timeZone string) *cron.Cron
	Push(c *cron.Cron, time string, call func()) (cron.EntryID, error)
	Start(c *cron.Cron)
	Stop(c *cron.Cron)
}

type CronProxyImpl struct {
}

func (i *CronProxyImpl) Parse(spec string) (cron.Schedule, error) {
	return V1Parser.Parse(spec)
}

func (i *CronProxyImpl) Create(timeZone string) *cron.Cron {
	l, _ := time.LoadLocation(timeZone)

	return cron.New(cron.WithLocation(l), cron.WithParser(V1Parser))
}

func (i *CronProxyImpl) Push(c *cron.Cron, time string, call func()) (cron.EntryID, error) {
	return c.AddFunc(time, call)
}

func (i *CronProxyImpl) Start(c *cron.Cron) {
	c.Start()
}

func (i *CronProxyImpl) Stop(c *cron.Cron) {
	c.Stop()
}
