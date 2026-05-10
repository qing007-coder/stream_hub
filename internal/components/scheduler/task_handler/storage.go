package task_handler

import (
	"context"
	"stream_hub/pkg/constant"
	infra_ "stream_hub/pkg/model/infra"
	"stream_hub/pkg/model/storage"

	"gorm.io/gorm"
)

func (t *TaskHandler) StorePersistency(ctx context.Context, task *infra_.TaskMessage) error {

	switch task.Payload.Action {
	case constant.ActionCreateLike:
		return t.createLike(task.Payload.Operator, task.BizID)
	case constant.ActionDeleteLike:
		return t.deleteLike(task.Payload.Operator, task.BizID)
	case constant.ActionCreateFavourite:
		return t.createFavourite(task.Payload.Operator, task.BizID)
	case constant.ActionDeleteFavourite:
		return t.deleteFavourite(task.Payload.Operator, task.BizID)
	}

	return nil
}

func (t *TaskHandler) createLike(userID, videoID string) error {
	if err := t.DB.Where("user_id = ? and video_id = ?", userID, videoID).FirstOrCreate(&storage.VideoLikeModel{
		UserID:  userID,
		VideoID: videoID,
	}).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	var video storage.VideoModel
	if err := t.DB.Select("author_id").Where("id = ?", videoID).First(&video).Error; err != nil {
		return err
	}

	if err := t.DB.Model(&storage.VideoModel{}).Where("id = ?", videoID).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
		return err
	}

	if err := t.DB.Model(&storage.User{}).Where("id = ?", video.AuthorID).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
		return err
	}

	return nil
}

func (t *TaskHandler) createFavourite(userID, videoID string) error {
	if err := t.DB.Where("user_id = ? and video_id = ?", userID, videoID).FirstOrCreate(&storage.VideoFavoriteModel{
		UserID:  userID,
		VideoID: videoID,
	}).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	var video storage.VideoModel
	if err := t.DB.Select("author_id").Where("id = ?", videoID).First(&video).Error; err != nil {
		return err
	}

	if err := t.DB.Model(&storage.VideoModel{}).Where("id = ?", videoID).UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1)).Error; err != nil {
		return err
	}

	if err := t.DB.Model(&storage.User{}).Where("id = ?", video.AuthorID).UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1)).Error; err != nil {
		return err
	}

	return nil
}

func (t *TaskHandler) deleteLike(userID, videoID string) error {
	if err := t.DB.Where("user_id = ? and video_id = ?", userID, videoID).Delete(&storage.VideoLikeModel{}).Error; err != nil {
		return err
	}

	var video storage.VideoModel
	if err := t.DB.Select("author_id").Where("id = ?", videoID).First(&video).Error; err != nil {
		return err
	}

	if err := t.DB.Model(&storage.VideoModel{}).Where("id = ?", videoID).UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
		return err
	}

	if err := t.DB.Model(&storage.User{}).Where("id = ?", video.AuthorID).UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
		return err
	}

	return nil
}

func (t *TaskHandler) deleteFavourite(userID, videoID string) error {
	if err := t.DB.Where("user_id = ? and video_id = ?", userID, videoID).Delete(&storage.VideoFavoriteModel{}).Error; err != nil {
		return err
	}

	var video storage.VideoModel
	if err := t.DB.Select("author_id").Where("id = ?", videoID).First(&video).Error; err != nil {
		return err
	}

	if err := t.DB.Model(&storage.VideoModel{}).Where("id = ?", videoID).UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1)).Error; err != nil {
		return err
	}

	if err := t.DB.Model(&storage.User{}).Where("id = ?", video.AuthorID).UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1)).Error; err != nil {
		return err
	}

	return nil
}