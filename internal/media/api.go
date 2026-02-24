package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"path"
	"sort"
	"strconv"
	"stream_hub/internal/infra"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/api"
	"stream_hub/pkg/model/config"
	infra_ "stream_hub/pkg/model/infra"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"
	"strings"
	"time"
)

type MediaApi struct {
	*infra.Base
	ChunkSize int
}

func NewMediaApi(base *infra.Base, conf *config.MediaConfig) *MediaApi {
	return &MediaApi{base, conf.ChunkSize}
}

func (m *MediaApi) UploadImage(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("image")
	imageType := ctx.PostForm("type")
	if err != nil {
		utils.BadRequest(ctx, "image is required")
		return
	}

	const maxSize = 5 << 20 // 5MB
	if fileHeader.Size > maxSize {
		utils.BadRequest(ctx, "image too large")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		utils.InternalServerError(ctx)
		return
	}
	defer file.Close()

	var bucketName string
	objectName := utils.CreateID() + path.Ext(fileHeader.Filename)

	switch imageType {
	case "private":
		bucketName = constant.PrivateImageBucket
	case "public":
		bucketName = constant.PublicImageBucket
	default:
		utils.BadRequest(ctx, "need image type")
		return
	}

	info, err := m.Minio.Client.PutObject(
		context.Background(),
		bucketName,
		objectName,
		file,
		fileHeader.Size,
		minio.PutObjectOptions{
			ContentType: fileHeader.Header.Get("Content-Type"),
		},
	)
	if err != nil {
		utils.InternalServerError(ctx)
		return
	}

	utils.StatusOK(ctx, gin.H{
		"object": fmt.Sprintf("/%s/%s", bucketName, info.Key),
		"size":   info.Size,
	}, "upload image success")
}

func (m *MediaApi) InitUpload(ctx *gin.Context) {
	var req api.InitUploadReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, "invalid request body")
		return
	}

	var video storage.FileModel
	if err := m.DB.Where("file_hash = ? and status = 2", req.FileHash).First(&video).Error; err == nil {
		utils.StatusOK(ctx, api.InitUploadResp{
			IsSkipped: true,
			VideoURL:  video.FilePath,
		}, "upload successfully")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.InternalServerError(ctx)
		return
	}

	key := fmt.Sprintf("upload:info:%s", req.FileHash)

	data, err := m.Redis.HGetAll(context.Background(), key)
	if err != nil {
		utils.InternalServerError(ctx)
		return
	}

	if len(data) != 0 {
		finish := make([]int, 0)
		for k, _ := range data {
			if !strings.HasPrefix(k, "part:") {
				continue
			}

			part, _ := strconv.Atoi(k[5:])
			finish = append(finish, part)
		}

		size, _ := strconv.Atoi(data["chunk_size"])

		utils.StatusOK(ctx, api.InitUploadResp{
			IsSkipped:   false,
			UploadID:    data["upload_id"],
			FinishChunk: finish,
			ChunkSize:   int64(size * 1024 * 1024),
		}, "find unfinished parts")

		return
	}

	fileName := m.GenerateObjectName(req.FileHash)

	id, err := m.Minio.Core.NewMultipartUpload(context.Background(), constant.VideoBucket, fileName, minio.PutObjectOptions{})
	if err != nil {
		utils.InternalServerError(ctx)
		return
	}

	info := map[string]interface{}{
		"upload_id":     id,
		"upload_chunks": 0,
		"file_name":     fileName,
		"file_size":     req.FileSize,
		"chunk_size":    m.ChunkSize,
	}
	if err := m.Redis.HSet(context.Background(), key, info); err != nil {
		utils.InternalServerError(ctx)
		return
	}

	if err := m.Redis.Expire(context.Background(), key, time.Hour*24); err != nil {
		utils.InternalServerError(ctx)
		return
	}

	m.DB.Create(&storage.FileModel{
		FileHash: req.FileHash,
		FilePath: fileName,
		Size:     req.FileSize,
		FileType: req.FileType,
		Status:   constant.FileStatusUploading,
	})

	utils.StatusOK(ctx, api.InitUploadResp{
		IsSkipped: false,
		UploadID:  id,
		ChunkSize: int64(m.ChunkSize * 1024 * 1024),
	}, "init successfully")
}

func (m *MediaApi) UploadChunk(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		utils.BadRequest(ctx, "file is required")
		return
	}

	fileHash := ctx.PostForm("file_hash")
	uploadID := ctx.PostForm("upload_id")
	partNumber, err := strconv.Atoi(ctx.PostForm("part_number"))
	if err != nil {
		utils.BadRequest(ctx, "part number is required")
		return
	}

	key := fmt.Sprintf("upload:info:%s", fileHash)
	record, err := m.Redis.HGetAll(context.Background(), key)
	if err != nil || len(record) == 0 {
		utils.BadRequest(ctx, "upload chunk is invalid")
		return
	}

	data, _ := file.Open()
	defer data.Close()

	part, err := m.Minio.Core.PutObjectPart(context.Background(), constant.VideoBucket, record["file_name"], uploadID, partNumber, data, file.Size, minio.PutObjectPartOptions{})
	if err != nil {
		utils.InternalServerError(ctx)
		return
	}

	partKey := fmt.Sprintf("part:%d", part.PartNumber)
	info, err := json.Marshal(part)
	if err != nil {
		utils.InternalServerError(ctx)
		return
	}

	pipe := m.Redis.Pipeline()
	pipe.HSet(context.Background(), key, map[string]interface{}{
		partKey: info,
	})
	pipe.HIncrBy(context.Background(), key, "upload_chunks", 1)
	_, err = pipe.Exec(context.Background())
	if err != nil {
		utils.InternalServerError(ctx)
		return
	}

	utils.StatusOK(ctx, api.UploadChunkResp{
		PartNumber: part.PartNumber,
		ETag:       part.ETag,
		Size:       part.Size,
	}, "upload chunk successfully")
}

func (m *MediaApi) CompleteUpload(ctx *gin.Context) {
	var req api.CompleteUploadReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, "invalid request body")
		return
	}

	key := fmt.Sprintf("upload:info:%s", req.FileHash)
	data, err := m.Redis.HGetAll(context.Background(), key)
	if err != nil || len(data) == 0 {
		utils.BadRequest(ctx, "not complete upload")
		return
	}

	parts := make([]minio.CompletePart, 0)
	for k, v := range data {
		if !strings.HasPrefix(k, "part:") {
			continue
		}

		var part minio.CompletePart
		if err := json.Unmarshal([]byte(v), &part); err != nil {
			utils.InternalServerError(ctx)
			return
		}

		parts = append(parts, part)
	}

	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})

	_, err = m.Minio.Core.CompleteMultipartUpload(context.Background(), constant.VideoBucket, data["file_name"], req.UploadID, parts, minio.PutObjectOptions{})
	if err != nil {
		utils.InternalServerError(ctx)
		return
	}

	m.DB.Model(&storage.FileModel{}).Where("file_hash = ?", req.FileHash).Update("status", constant.FileStatusUploadFinished)

	var video storage.FileModel
	m.DB.Where("file_hash = ?", req.FileHash).First(&video)

	// 发送转码任务
	if err := m.TaskSender.SendTask(infra_.TaskMessage{
		Type:    constant.TaskVideoTranscode,
		BizID:   video.ID,
		Priority: "critical",
		RetryCount: 0,
		Payload: infra_.TaskPayload{
			Operator: "",
			Source: constant.Media,
			Data: nil,
		},
	}); err != nil {
		utils.InternalServerError(ctx)
		return
	}

	utils.StatusOK(ctx, api.CompleteUploadResp{}, "finish uploading successfully")
}

func (m *MediaApi) GenerateObjectName(fileName string) string {
	firstDirection := fileName[0:2]
	secondDirection := fileName[2:4]

	return fmt.Sprintf("%s/%s/%s.mp4", firstDirection, secondDirection, fileName)
}
