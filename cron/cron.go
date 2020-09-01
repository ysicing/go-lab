// MIT License
// Copyright (c) 2020 ysicing <i@ysicing.me>

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/ysicing/logger"
)

func init()  {
	cfg := logger.LogConfig{Simple: true}
	logger.InitLogger(&cfg)
}

func main()  {
	r := gin.New()
	cronjob := CronJob{Cron: cron.New()}
	cronjob.Start()
	defer cronjob.Stop()
	r.Run(":65535")
}