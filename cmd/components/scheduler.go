package main

import (
	"fmt"
	"stream_hub/internal/components/scheduler/core"
	"stream_hub/internal/components/scheduler/task_handler"
	"stream_hub/internal/infra"
	"stream_hub/pkg/config"
	"stream_hub/pkg/constant"
)

func main() {
	commonConf, err := config.NewCommonConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	schedulerConf, err := config.NewSchedulerConfig()
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	base, err := infra.NewBase(commonConf)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	handler := task_handler.NewCommonTaskHandler(commonConf, base)

	server := core.NewServer(base.DB, base.Redis, schedulerConf)

	serveMux := core.NewServeMux()
	serveMux.HandleFunc(constant.TaskSendEmailCode, handler.EmailHandler)
	serveMux.HandleFunc(constant.TaskVideoTranscode, handler.TranscodeHandler)

	server.RegisterServeMux(serveMux)

	deadletter := core.NewDeadLetter(base.DB, base.Redis, schedulerConf)
	dispatcher := core.NewDispatcher(base.Redis, schedulerConf)
	janitor := core.NewJanitor(base.Redis, schedulerConf)

	go deadletter.Start()
	go dispatcher.Start()
	go janitor.Run()

	if err := server.Start(); err != nil {
		fmt.Println("err:", err)
	}
}
