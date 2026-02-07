package interaction

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"stream_hub/internal/infra"
	pb "stream_hub/internal/proto/interaction"
	"stream_hub/pkg/model/storage"
	"time"
)

type Favourite struct {
	*infra.Base
}

func NewFavourite(base *infra.Base) *Favourite {
	return &Favourite{
		base,
	}
}

func (f *Favourite) CreateFavorite(ctx context.Context, req *pb.FavoriteRequest, resp *pb.ActionResponse) error {
	uid := ctx.Value("user_id").(string)
	if req.VideoId == "" {
		return errors.New("video_id is empty")
	}

	pipeline := f.Redis.Pipeline()
	pipeline.Incr(ctx, fmt.Sprintf("video:favorite:count:%s", req.VideoId))
	pipeline.ZAdd(ctx, fmt.Sprintf("user:favorite:video:%s", uid), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: req.VideoId,
	})
	pipeline.ZAdd(ctx, fmt.Sprintf("video:favorite:user:%s", req.VideoId), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: uid,
	})
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return err
	}

	if err := f.DB.Create(&storage.VideoFavoriteModel{
		UserID:  uid,
		VideoID: req.VideoId,
	}).Error; err != nil {
		return err
	}

	resp.Success = true
	resp.Message = "ok"
	return nil
}

func (f *Favourite) DeleteFavorite(ctx context.Context, req *pb.FavoriteRequest, resp *pb.ActionResponse) error {
	uid := ctx.Value("user_id").(string)
	if req.VideoId == "" {
		return errors.New("video_id is empty")
	}

	pipeline := f.Redis.Pipeline()
	pipeline.Decr(ctx, fmt.Sprintf("video:favorite:count:%s", req.VideoId))
	pipeline.ZRem(ctx, fmt.Sprintf("user:favorite:video:%s", uid), req.VideoId)
	pipeline.ZRem(ctx, fmt.Sprintf("video:favorite:user:%s", req.VideoId), uid)
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return err
	}

	if err := f.DB.Where("user_id = ? and video_id = ?", uid, req.VideoId).Delete(&storage.VideoFavoriteModel{}).Error; err != nil {
		return err
	}

	resp.Success = true
	resp.Message = "ok"
	return nil
}

func (f *Favourite) IsFavorite(ctx context.Context, req *pb.IsFavoriteRequest, resp *pb.IsFavoriteResponse) error {
	uid := ctx.Value("user_id").(string)

	_, err := f.Redis.ZScore(ctx, fmt.Sprintf("user:favorite:video:%s", uid), req.VideoId)
	if err == nil {
		resp.IsFavorite = true
		return nil
	}

	if !errors.Is(err, redis.Nil) {
		return err
	}

	var count int64
	f.DB.Model(&storage.VideoFavoriteModel{}).Where("user_id = ? and video_id = ?", uid, req.VideoId).Count(&count)
	if count == 0 {
		resp.IsFavorite = false
		return nil
	}

	_ = f.Redis.ZAdd(ctx, fmt.Sprintf("user:favorite:video:%s", uid), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: req.VideoId,
	})

	resp.IsFavorite = true
	return nil
}
