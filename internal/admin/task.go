package admin

import (
	"fmt"
	"stream_hub/internal/infra"
	"stream_hub/pkg/model/api"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"

	"github.com/gin-gonic/gin"
)

type TaskController struct {
	*infra.Base
}

func NewTaskController(base *infra.Base) *TaskController {
	return &TaskController{
		base,
	}
}

func (t *TaskController) GetTaskList(ctx *gin.Context) {
	var req api.GetTaskListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(ctx, "invalid param")
		return
	}

	var tasks []storage.Task
	var total int64

	offset := (req.Page - 1) * req.Size
	query := t.DB.Model(&storage.Task{}).Offset(offset).Limit(req.Size)

	if req.TaskType != "" {
		query = query.Where("type = ?", req.TaskType)
	}

	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if req.BizID != "" {
		query = query.Where("biz_id = ?", req.BizID)
	}

	if err := query.Count(&total).Error; err != nil {
		utils.BadRequest(ctx, fmt.Sprintf("count error: %s", err.Error()))
		return
	}

	if err := query.Order("created_at DESC").Find(&tasks).Error; err != nil {
		utils.BadRequest(ctx, fmt.Sprintf("query error: %s", err.Error()))
		return
	}

	utils.StatusOK(ctx, gin.H{
		"total": total,
		"list":  tasks,
	}, "查询成功")
}

func (t *TaskController) GetTaskDetail(ctx *gin.Context) {
	id := ctx.Param("id")

	var task storage.Task
	if err := t.DB.Where("id = ?", id).First(&task).Error; err != nil {
		utils.BadRequest(ctx, fmt.Sprintf("task not found: %s", err.Error()))
		return
	}

	utils.StatusOK(ctx, gin.H{
		"id":          task.ID,
		"type":        task.Type,
		"biz_id":      task.BizID,
		"status":      task.Status,
		"retry_count": task.RetryCount,
		"error_msg":   task.ErrorMsg,
		"payload":     task.Payload,
		"next_run_at": task.NextRunAt,
		"executor":    task.Executor,
		"remark":      task.Remark,
		"created_at":  task.CreatedAt,
		"updated_at":  task.UpdatedAt,
	}, "查询成功")
}

func (t *TaskController) UpdateTaskStatus(ctx *gin.Context) {
	var req api.UpdateTaskStatusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, "invalid request body")
		return
	}

	var task storage.Task
	if err := t.DB.Where("id = ?", req.TaskID).First(&task).Error; err != nil {
		utils.BadRequest(ctx, fmt.Sprintf("task not found: %s", err.Error()))
		return
	}

	if err := t.DB.Model(&storage.Task{}).Where("id = ?", req.TaskID).Update("status", req.Status).Error; err != nil {
		utils.BadRequest(ctx, fmt.Sprintf("update error: %s", err.Error()))
		return
	}

	utils.StatusOK(ctx, gin.H{
		"task_id": req.TaskID,
		"status":  req.Status,
	}, "更新成功")
}
