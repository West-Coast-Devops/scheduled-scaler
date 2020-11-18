package cron

//go:generate mockgen -source $GOFILE -destination=mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE

import (
	"github.com/robfig/cron"
	"time"
)

// CronProxy wraps the cron object for testing purposes, as this interface can be mocked.
type CronProxy interface {
	Parse(spec string) (cron.Schedule, error)
	Create(timeZone string) (*cron.Cron, error)
	Push(c *cron.Cron, time string, call func())
	Start(c *cron.Cron)
	Stop(c *cron.Cron)
}

// CronImpl passes methods through to cron methods.
type CronImpl struct {
}

// Parse parses the spec into a schedule.
func (ci *CronImpl) Parse(spec string) (cron.Schedule, error) {
	return cron.Parse(spec)
}

// Create creates a cron object for the given timeZone.
func (ci *CronImpl) Create(timeZone string) (*cron.Cron, error) {
	l, err := time.LoadLocation(timeZone)
	if err != nil {
		return nil, err
	}

	return cron.NewWithLocation(l), nil
}

// Push pushes the time spec onto the cron, c, with call callback.
func (ci *CronImpl) Push(c *cron.Cron, time string, call func()) {
	s, _ := cron.Parse(time)
	c.Schedule(s, cron.FuncJob(call))
}

// Start starts the cron object, c.
func (ci *CronImpl) Start(c *cron.Cron) {
	c.Start()
}

// Stop stops the cron object, c.
func (ci *CronImpl) Stop(c *cron.Cron) {
	c.Stop()
}
