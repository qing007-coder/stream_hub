package interaction

import (
	"context"
	"errors"
	"fmt"
	"stream_hub/internal/infra"
	pb "stream_hub/internal/proto/interaction"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"
	"time"

	"github.com/go-redis/redis/v8"
)

type Like struct {
	*infra.Base
	sender *EventSender
}

func NewLike(base *infra.Base, sender *EventSender) *Like {
	return &Like{
		base,
		sender,
	}
}

func (l *Like) CreateLike(ctx context.Context, req *pb.LikeRequest, resp *pb.ActionResponse) error {
	uid := ctx.Value("user_id").(string)

	if req.VideoId == "" {
		return errors.New("video_id is empty")
	}

	pipeline := l.Redis.Pipeline()
	pipeline.Incr(ctx, fmt.Sprintf("video:like:count:%s", req.VideoId))
	pipeline.ZAdd(ctx, fmt.Sprintf("video:like:user:%s", req.VideoId), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: uid,
	})
	pipeline.ZAdd(ctx, fmt.Sprintf("user:like:video:%s", uid), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: req.VideoId,
	})
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return err
	}

	if err := l.DB.Create(&storage.VideoLikeModel{
		UserID:  uid,
		VideoID: req.VideoId,
	}).Error; err != nil {
		return err
	}

	eventType := ctx.Value("event_type").(string)

	l.sender.Send(&storage.Event{
		EventID:      utils.CreateID(),
		EventType:    eventType,
		ResourceType: constant.ResourceVideo,
		ResourceID:   req.VideoId,
		Timestamp:    time.Now().Unix(),
	})

	resp.Success = true
	resp.Message = "ok"
	return nil
}

func (l *Like) DeleteLike(ctx context.Context, req *pb.LikeRequest, resp *pb.ActionResponse) error {
	uid := ctx.Value("user_id").(string)

	if req.VideoId == "" {
		return errors.New("video_id is empty")
	}

	pipeline := l.Redis.Pipeline()
	pipeline.Decr(ctx, fmt.Sprintf("video:like:count:%s", req.VideoId))
	pipeline.ZRem(ctx, fmt.Sprintf("video:like:user:%s", req.VideoId), uid)
	pipeline.ZRem(ctx, fmt.Sprintf("user:like:video:%s", uid), req.VideoId)
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return err
	}

	if err := l.DB.Where("user_id = ? and video_id = ?", uid, req.VideoId).Delete(&storage.VideoLikeModel{}).Error; err != nil {
		return err
	}

	resp.Success = true
	resp.Message = "ok"
	return nil
}

func (l *Like) IsLike(ctx context.Context, req *pb.IsLikeRequest, resp *pb.IsLikeResponse) error {
	uid := ctx.Value("user_id").(string)

	_, err := l.Redis.ZScore(ctx, fmt.Sprintf("video:like:user:%s", req.VideoId), uid)
	if err == nil {
		resp.IsLike = true
		return nil
	}

	if !errors.Is(err, redis.Nil) {
		return err
	}

	var count int64
	l.DB.Model(&storage.VideoLikeModel{}).Where("user_id = ? and video_id = ?", uid, req.VideoId).Count(&count)
	if count == 0 {
		resp.IsLike = false
		return nil
	}

	if err := l.Redis.ZAdd(ctx, fmt.Sprintf("video:like:user:%s", req.VideoId), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: uid,
	}); err != nil {
		return err
	}

	resp.IsLike = true
	return nil
}

func (l *Like) ListLikes(ctx context.Context, req *pb.ListLikesRequest, resp *pb.ListLikesResponse) error {
	if req.VideoId == "" {
		return errors.New("video_id is empty")
	}

	key := fmt.Sprintf("video:like:user:%s", req.VideoId)

	total, err := l.Redis.ZCard(ctx, key)
	if err != nil {
		return err
	}

	if total == 0 {
		return nil
	}

	resp.Total = total

	start := (req.Page - 1) * req.Size
	stop := start + req.Size - 1
	res, err := l.Redis.ZRangeArgsWithScores(ctx, redis.ZRangeArgs{
		Key:   key,
		Start: start,
		Stop:  stop,
		Rev:   true, // 降序排列，适合排行榜
	})
	if err != nil {
		return err
	}

	ids := make([]string, 0)
	for _, z := range res {
		ids = append(ids, z.Member.(string))
	}

	var users []storage.User
	l.DB.Where("id in (?)", ids).Find(&users)

	for _, user := range users {
		resp.Users = append(resp.Users, &pb.UserView{
			UserId:    user.ID,
			Nickname:  user.Nickname,
			Avatar:    user.Avatar,
			Signature: user.Signature,
		})
	}

	return nil
}
