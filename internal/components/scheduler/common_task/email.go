package common_task

import "stream_hub/pkg/model/infra"

type EmailWorker struct {
	taskCh <-chan infra.TaskMessage
}
