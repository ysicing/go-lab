// MIT License
// Copyright (c) 2020 ysicing <i@ysicing.me>

package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/ysicing/go-utils/extime"
	"github.com/ysicing/logger"
)

type CronJob struct {
	Cron *cron.Cron
}

func (cj *CronJob) Start()  {
	logger.Debug("start cron job")
	cj.Cron.AddFunc("@every 1s", func() {
		logger.Debug(fmt.Sprintf("now: %v", extime.NowUnixString()))
	})
	cj.Cron.Start()
}

func (cj *CronJob) Stop() {
	defer cj.Cron.Stop()
}