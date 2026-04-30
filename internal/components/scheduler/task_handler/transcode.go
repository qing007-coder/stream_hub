package task_handler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"stream_hub/pkg/constant"
	infra_ "stream_hub/pkg/model/infra"
	"stream_hub/pkg/model/storage"
	"time"

	"github.com/minio/minio-go/v7"
)

func (t *TaskHandler) TranscodeHandler(ctx context.Context, task *infra_.TaskMessage) error {
	var media storage.FileModel
	if err := t.DB.Where("id = ?", task.BizID).First(&media).Error; err != nil {
		return err
	}

	if media.Status == constant.FileStatusTranscodeFinished {
		return nil
	}

	localTmpDir := filepath.Join("./tmp", media.ID) // 本地临时存放切片的目录

	// 确保本地目录存在，处理完后自动清理
	if err := os.MkdirAll(localTmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create tmp dir: %w", err)
	}
	defer os.RemoveAll(localTmpDir)

	// 生成 MinIO 临时下载链接 (让 FFmpeg 能够读取私有桶文件)
	expiry := time.Hour * 2
	presignedURL, err := t.Minio.Client.PresignedGetObject(ctx, constant.VideoBucket, media.FilePath, expiry, nil)
	if err != nil {
		return fmt.Errorf("failed to generate presigned url: %w", err)
	}

	// 构造 FFmpeg 命令转码为 HLS (m3u8 + ts)
	m3u8Path := filepath.Join(localTmpDir, "index.m3u8")
	// %03d.ts 会生成 seg001.ts, seg002.ts 等
	segmentPath := filepath.Join(localTmpDir, "seg%03d.ts")

	args := []string{
		"-i", presignedURL.String(), // 输入：MinIO 临时链接
		"-c:v", "libx264", // 视频编码
		"-c:a", "aac", // 音频编码
		"-f", "hls", // 输出格式为 HLS
		"-hls_time", "10", // 每个切片 10 秒
		"-hls_list_size", "0", // 索引保留所有切片
		"-hls_segment_filename", segmentPath,
		m3u8Path,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg transcode failed: %w", err)
	}

	// 批量上传转码后的文件到 MinIO
	files, _ := os.ReadDir(localTmpDir)
	for _, file := range files {
		localFile := filepath.Join(localTmpDir, file.Name())
		// 上传到 MinIO 的路径，比如：output/video_123/index.m3u8
		targetKey := fmt.Sprintf("output/%s/%s", media.ID, file.Name())

		_, err := t.Minio.Client.FPutObject(ctx, constant.VideoBucket, targetKey, localFile, minio.PutObjectOptions{
			ContentType: t.getContentType(file.Name()), // 根据后缀设置类型
		})
		if err != nil {
			return fmt.Errorf("failed to upload %s: %w", file.Name(), err)
		}
	}

	// 更新数据库状态，标记转码完成
	if err := t.DB.Model(&storage.FileModel{}).Where("id = ?", media.ID).Update("status", constant.FileStatusTranscodeFinished).Error; err != nil {
		return err
	}

	t.DB.Model(&storage.MediaModel{}).Where("id = ?", task.BizID).Updates(map[string]interface{}{
		"transcode_status": constant.FileStatusTranscodeFinished,
		"m3u8_url": fmt.Sprintf("/%s/output/%s/index.m3u8", constant.VideoBucket, media.ID),
	})


	return t.TaskSender.SendTask(infra_.TaskMessage{
		Type: constant.TaskVideoAudit,
		BizID: task.BizID,
		Priority: "critical",
		RetryCount: 0,
		Payload: infra_.TaskPayload{
			Operator: "",
			Action: "",
			Source: constant.Scheduler,
			Data: nil,
		},
	})
}

func (t *TaskHandler) getContentType(fileName string) string {
	switch filepath.Ext(fileName) {
	case ".m3u8":
		return "application/x-mpegURL"
	case ".ts":
		return "video/MP2T"
	default:
		return "application/octet-stream"
	}
}
