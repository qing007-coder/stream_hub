package task_handler

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"stream_hub/pkg/constant"
	infra_ "stream_hub/pkg/model/infra"
	"stream_hub/pkg/model/storage"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

func (t *TaskHandler) TranscodeHandler(ctx context.Context, task *infra_.TaskMessage) error {
	var asset storage.MediaModel
	if err := t.DB.Where("id = ?", task.BizID).First(&asset).Error; err != nil {
		return err
	}

	var media storage.FileModel
	if err := t.DB.Where("file_path = ?", asset.SourceObjectKey).First(&media).Error; err != nil {
		return err
	}

	// 本地临时存放切片的目录
	localTmpDir := filepath.Join("./tmp", media.ID)

	// 确保本地目录存在，处理完后自动清理
	if err := os.MkdirAll(localTmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create tmp dir: %w", err)
	}
	defer os.RemoveAll(localTmpDir)

	// 从数据库中的文件路径中去掉 bucket 前缀，得到 MinIO objectName
	bucketPrefix := "/" + constant.VideoBucket + "/"
	objectName := strings.TrimPrefix(media.FilePath, bucketPrefix)

	// 生成 MinIO 临时下载链接，让 ffmpeg / ffprobe 可以读取私有桶文件
	expiry := time.Hour * 2
	presignedURL, err := t.Minio.Client.PresignedGetObject(
		ctx,
		constant.VideoBucket,
		objectName,
		expiry,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to generate presigned url: %w", err)
	}

	// 获取源视频时长，单位：秒，整型，向上取整
	durationSec, err := t.probeVideoDurationSeconds(ctx, presignedURL.String())
	if err != nil {
		return fmt.Errorf("failed to probe video duration: %w", err)
	}

	// 构造 FFmpeg 命令转码为 HLS
	m3u8Path := filepath.Join(localTmpDir, "index.m3u8")
	segmentPath := filepath.Join(localTmpDir, "seg%03d.ts")

	args := []string{
		"-hide_banner",
		"-y",
		"-i", presignedURL.String(),

		// 映射视频流和可选音频流
		"-map", "0:v:0",
		"-map", "0:a?",

		// 视频不重新编码，直接封装为 HLS
		"-c:v", "copy",

		// 音频转为 AAC，保证 HLS 兼容性
		"-c:a", "aac",

		// 输出 HLS
		"-f", "hls",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", segmentPath,

		m3u8Path,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg transcode failed: %w\nffmpeg output:\n%s", err, string(output))
	}

	// 读取本地 HLS 文件
	files, err := os.ReadDir(localTmpDir)
	if err != nil {
		return fmt.Errorf("failed to read tmp dir: %w", err)
	}

	// 批量上传转码后的文件到 MinIO
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		localFile := filepath.Join(localTmpDir, file.Name())

		// 上传到 MinIO 的路径，例如：output/video_123/index.m3u8
		targetKey := fmt.Sprintf("output/%s/%s", media.ID, file.Name())

		_, err := t.Minio.Client.FPutObject(ctx, constant.VideoBucket, targetKey, localFile, minio.PutObjectOptions{
			ContentType: t.getContentType(file.Name()),
		})
		if err != nil {
			return fmt.Errorf("failed to upload %s: %w", file.Name(), err)
		}
	}

	// 更新文件状态
	if err := t.DB.Model(&storage.FileModel{}).
		Where("id = ?", media.ID).
		Update("status", constant.FileStatusTranscodeFinished).Error; err != nil {
		return err
	}

	// 更新媒体转码状态、m3u8 地址
	if err := t.DB.Model(&storage.MediaModel{}).
		Where("id = ?", task.BizID).
		Updates(map[string]interface{}{
			"transcode_status": constant.FileStatusTranscodeFinished,
			"m3u8_url":         fmt.Sprintf("/%s/output/%s/index.m3u8", constant.VideoBucket, media.ID),
		}).Error; err != nil {
		return err
	}

	// 更新视频时长到 VideoModel
	if err := t.DB.Model(&storage.VideoModel{}).
		Where("id = ?", asset.VideoID).
		Update("duration", durationSec).Error; err != nil {
		return err
	}

	var video storage.VideoModel
	if err := t.DB.Select("author_id", "title").Where("id = ?", asset.VideoID).First(&video).Error; err == nil && video.AuthorID != "" {
		_ = t.SendNotify(ctx, &NotifyPayload{
			ReceiverID:  video.AuthorID,
			SenderID:    constant.IMSystemSenderID,
			Content:     "你的视频《" + video.Title + "》已完成转码，正在机审",
			ContentType: constant.IMContentTypeSystem,
		})
	}

	// 发送视频审核任务
	return t.TaskSender.SendTask(infra_.TaskMessage{
		Type:       constant.TaskVideoAudit,
		BizID:      task.BizID,
		Priority:   "critical",
		RetryCount: 0,
		Payload: infra_.TaskPayload{
			Operator: "",
			Action:   "",
			Source:   constant.Scheduler,
			Data:     nil,
		},
	})
}

// probeVideoDurationSeconds 获取视频时长
//
// 返回值：
// - 单位：秒
// - 类型：int64
// - 规则：向上取整，例如 12.1 秒会返回 13
func (t *TaskHandler) probeVideoDurationSeconds(ctx context.Context, inputURL string) (int64, error) {
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputURL,
	}

	cmd := exec.CommandContext(ctx, "ffprobe", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w\nffprobe output:\n%s", err, string(output))
	}

	durationText := strings.TrimSpace(string(output))
	if durationText == "" || durationText == "N/A" {
		return 0, fmt.Errorf("video duration is empty or N/A")
	}

	durationFloat, err := strconv.ParseFloat(durationText, 64)
	if err != nil {
		return 0, fmt.Errorf("parse video duration failed: %w, raw=%s", err, durationText)
	}

	if durationFloat <= 0 {
		return 0, fmt.Errorf("invalid video duration: %f", durationFloat)
	}

	durationSec := int64(math.Ceil(durationFloat))

	return durationSec, nil
}

// getContentType 根据文件后缀设置上传到 MinIO 的 Content-Type
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
