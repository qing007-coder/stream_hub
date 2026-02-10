package interaction

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"stream_hub/internal/infra"
	pb "stream_hub/internal/proto/interaction"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"
	"time"
)

type Follow struct {
	*infra.Base
	sender *EventSender
}

func NewFollow(base *infra.Base, sender *EventSender) *Follow {
	return &Follow{base, sender}
}

// CreateFollow 关注
func (f *Follow) CreateFollow(ctx context.Context, req *pb.FollowRequest, resp *pb.ActionResponse) error {
	uid := ctx.Value("user_id").(string)
	if req.TargetUserId == "" {
		return errors.New("target_user_id is empty")
	}
	if uid == req.TargetUserId {
		return errors.New("cannot follow yourself")
	}

	if err := f.DB.Create(&storage.UserFollowModel{
		UserID:       uid,
		TargetUserID: req.TargetUserId,
	}).Error; err != nil {
		return err
	}

	pipe := f.Redis.Pipeline()
	now := float64(time.Now().Unix())
	pipe.ZAdd(ctx, fmt.Sprintf("user:following:%s", uid), &redis.Z{Score: now, Member: req.TargetUserId})
	pipe.ZAdd(ctx, fmt.Sprintf("user:follower:%s", req.TargetUserId), &redis.Z{Score: now, Member: uid})
	_, err := pipe.Exec(ctx)

	eventType := ctx.Value("event_type").(string)

	f.sender.Send(&storage.Event{
		EventID:      utils.CreateID(),
		EventType:    eventType,
		ResourceType: constant.ResourceUser,
		ResourceID:   req.TargetUserId,
		Timestamp:    time.Now().Unix(),
	})

	resp.Success = true
	resp.Message = "ok"
	return err
}

// DeleteFollow 取消关注
func (f *Follow) DeleteFollow(ctx context.Context, req *pb.FollowRequest, resp *pb.ActionResponse) error {
	uid := ctx.Value("user_id").(string)

	if err := f.DB.Where("user_id = ? AND target_user_id = ?", uid, req.TargetUserId).Delete(&storage.UserFollowModel{}).Error; err != nil {
		return err
	}

	pipe := f.Redis.Pipeline()
	pipe.ZRem(ctx, fmt.Sprintf("user:following:%s", uid), req.TargetUserId)
	pipe.ZRem(ctx, fmt.Sprintf("user:follower:%s", req.TargetUserId), uid)
	_, err := pipe.Exec(ctx)

	eventType := ctx.Value("event_type").(string)

	f.sender.Send(&storage.Event{
		EventID:      utils.CreateID(),
		EventType:    eventType,
		ResourceType: constant.ResourceUser,
		ResourceID:   req.TargetUserId,
		Timestamp:    time.Now().Unix(),
	})

	resp.Success = true
	resp.Message = "ok"
	return err
}

// IsFollow 是否关注
func (f *Follow) IsFollow(ctx context.Context, req *pb.IsFollowRequest, resp *pb.IsFollowResponse) error {
	uid := ctx.Value("user_id").(string)
	_, err := f.Redis.ZScore(ctx, fmt.Sprintf("user:following:%s", uid), req.TargetUserId)
	if err == nil {
		resp.IsFollow = true
		return nil
	}

	// 如果 Redis 没命中，查 DB 并回填
	var count int64
	f.DB.Model(&storage.UserFollowModel{}).Where("user_id = ? AND target_user_id = ?", uid, req.TargetUserId).Count(&count)
	if count > 0 {
		resp.IsFollow = true
		_ = f.Redis.ZAdd(ctx, fmt.Sprintf("user:following:%s", uid), &redis.Z{Score: float64(time.Now().Unix()), Member: req.TargetUserId})
	} else {
		resp.IsFollow = false
	}
	return nil
}

// ListFollowers 粉丝列表 - 返回 UserView 切片
func (f *Follow) ListFollowers(ctx context.Context, req *pb.ListFollowersRequest, resp *pb.ListFollowersResponse) error {
	key := fmt.Sprintf("user:follower:%s", req.UserId)
	users, total, err := f.getDetailedUserList(ctx, key, req.Page, req.Size)
	if err != nil {
		return err
	}

	resp.Users = users
	resp.Total = total
	return nil
}

// ListFollowings 关注列表 - 返回 UserView 切片
func (f *Follow) ListFollowings(ctx context.Context, req *pb.ListFollowingsRequest, resp *pb.ListFollowingsResponse) error {
	key := fmt.Sprintf("user:following:%s", req.UserId)
	users, total, err := f.getDetailedUserList(ctx, key, req.Page, req.Size)
	if err != nil {
		return err
	}

	resp.Users = users
	resp.Total = total
	return nil
}

// getDetailedUserList 核心逻辑：从 Redis 分页拿 ID，再从 DB 批量补全 UserView 信息
func (f *Follow) getDetailedUserList(ctx context.Context, key string, page, size int32) ([]*pb.UserView, int64, error) {
	total, err := f.Redis.ZCard(ctx, key)
	if err != nil || total == 0 {
		return nil, total, err
	}

	start := int64((page - 1) * size)
	stop := start + int64(size) - 1
	ids, err := f.Redis.ZRevRange(ctx, key, start, stop)
	if err != nil {
		return nil, total, err
	}

	var storageUsers []storage.User
	if err := f.DB.Where("id IN ?", ids).Find(&storageUsers).Error; err != nil {
		return nil, total, err
	}

	userMap := make(map[string]*storage.User)
	for idx := range storageUsers {
		userMap[storageUsers[idx].ID] = &storageUsers[idx]
	}

	res := make([]*pb.UserView, 0, len(ids))
	for _, id := range ids {
		if u, ok := userMap[id]; ok {
			res = append(res, &pb.UserView{
				UserId:    u.ID,
				Nickname:  u.Nickname,
				Avatar:    u.Avatar,
				Signature: u.Signature,
			})
		}
	}

	return res, total, nil
}
