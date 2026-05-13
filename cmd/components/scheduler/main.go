package main

import (
	"fmt"
	"stream_hub/internal/components/scheduler/core"
	"stream_hub/internal/components/scheduler/task_handler"
	"stream_hub/internal/infra"
	"stream_hub/pkg/config"
	"stream_hub/pkg/constant"
	infra_ "stream_hub/pkg/model/infra"
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

	handler, err := task_handler.NewTaskHandler(commonConf, base)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	server := core.NewServer(base.DB, base.Redis, schedulerConf)

	serveMux := core.NewServeMux()
	serveMux.HandleFunc(constant.TaskSendEmailCode, handler.EmailHandler)
	serveMux.HandleFunc(constant.TaskVideoTranscode, handler.TranscodeHandler)
	serveMux.HandleFunc(constant.TaskCalculateFeed, handler.CalculateFeed)
	serveMux.HandleFunc(constant.TaskVideoAudit, handler.AuditVideo)
	serveMux.HandleFunc(constant.TaskStorePersistency, handler.StorePersistency)
	serveMux.HandleFunc(constant.TaskSendNotify, handler.NotifyTask)

	taskSender, err := infra.NewTaskSender(commonConf)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	// 初始化定时任务种子
	taskSender.SendTask(infra_.TaskMessage{
		Type:     constant.TaskCalculateFeed,
		BizID:    "",
		Priority: "critical",
		Payload: infra_.TaskPayload{
			Operator: "",
			Action:   "",
			Source:   constant.Scheduler,
			Data:     nil,
		},
	})

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
