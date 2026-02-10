package interaction

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"stream_hub/internal/infra"
	pb "stream_hub/internal/proto/interaction"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"
	"time"
)

type Comment struct {
	*infra.Base
	sender *EventSender
}

func NewComment(base *infra.Base, sender *EventSender) *Comment {
	return &Comment{base, sender}
}

func (c *Comment) CreateComment(ctx context.Context, req *pb.CreateCommentRequest, resp *pb.Comment) error {
	uid, ok := ctx.Value("user_id").(string)
	if !ok || uid == "" {
		return errors.New("unauthorized")
	}

	if req.VideoId == "" || req.Content == "" {
		return errors.New("invalid comment params")
	}

	now := time.Now().Unix()
	var user storage.User
	c.DB.Where("id = ? ", uid).First(&user)

	doc := storage.CommentModel{
		VideoID:       req.VideoId,
		UserID:        uid,
		Avatar:        user.Avatar,
		Nickname:      user.Nickname,
		Content:       req.Content,
		ParentID:      req.ParentId,
		ReplyToUserID: req.ReplyToUserId,
		LikeCount:     0,
		ReplyCount:    0,
		CreateTime:    now,
		IsDeleted:     false,
	}

	collection := c.Mongo.Collection(constant.InteractionDB, constant.Comment)

	res, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}

	oid := res.InsertedID.(primitive.ObjectID)

	eventType := ctx.Value("event_type").(string)

	c.sender.Send(&storage.Event{
		EventID:      utils.CreateID(),
		EventType:    eventType,
		ResourceType: constant.ResourceVideo,
		ResourceID:   req.VideoId,
		Timestamp:    time.Now().Unix(),
	})

	// 回填 response
	resp.Id = oid.Hex()
	resp.VideoId = doc.VideoID
	resp.UserId = doc.UserID
	resp.Content = doc.Content
	resp.ParentId = doc.ParentID
	resp.ReplyToUserId = doc.ReplyToUserID
	resp.LikeCount = 0
	resp.ReplyCount = 0
	resp.CreateTime = now

	return nil
}

func (c *Comment) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest, resp *pb.ActionResponse) error {

	uid, ok := ctx.Value("user_id").(string)
	if !ok || uid == "" {
		return errors.New("unauthorized")
	}

	if req.CommentId == "" {
		return errors.New("comment_id is empty")
	}

	oid, err := primitive.ObjectIDFromHex(req.CommentId)
	if err != nil {
		return errors.New("invalid comment_id")
	}

	collection := c.Mongo.Collection(constant.InteractionDB, constant.Comment)

	filter := bson.M{
		"_id":        oid,
		"user_id":    uid,
		"is_deleted": false,
	}

	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("comment not found or no permission")
	}

	resp.Success = true
	resp.Message = "ok"
	return nil
}

func (c *Comment) ListComments(ctx context.Context, req *pb.ListCommentsRequest, resp *pb.ListCommentsResponse) error {

	if req.VideoId == "" {
		return errors.New("video_id is empty")
	}

	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	skip := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	filter := bson.M{
		"video_id":   req.VideoId,
		"parent_id":  req.ParentId,
		"is_deleted": false,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "create_time", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	collection := c.Mongo.Collection(constant.InteractionDB, constant.Comment)

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	resp.Comments = make([]*pb.Comment, 0, req.PageSize)

	for cursor.Next(ctx) {
		var doc storage.CommentModel
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		resp.Comments = append(resp.Comments, &pb.Comment{
			Id:            doc.ID.Hex(),
			VideoId:       doc.VideoID,
			UserId:        doc.UserID,
			Content:       doc.Content,
			ParentId:      doc.ParentID,
			ReplyToUserId: doc.ReplyToUserID,
			LikeCount:     doc.LikeCount,
			ReplyCount:    doc.ReplyCount,
			CreateTime:    doc.CreateTime,
		})
	}

	// 是否还有更多
	resp.Total = int64(len(resp.Comments))
	return nil
}
