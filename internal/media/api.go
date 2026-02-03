package media

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"path"
	"stream_hub/internal/infra"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/utils"
)

type MediaApi struct {
	*infra.Base
}

func NewMediaApi(base *infra.Base) *MediaApi {
	return &MediaApi{base}
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

	info, err := m.Minio.PutObject(
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
