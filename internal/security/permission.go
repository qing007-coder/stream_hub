package security

import (
	"errors"
	"fmt"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/storage"

	"gorm.io/gorm"
)

type Permission struct {
	db *gorm.DB
}

func NewPermission(db *gorm.DB) *Permission {
	return &Permission{
		db: db,
	}
}

func (p *Permission) Enforcer(userID, resource, action string) (bool, error) {
	var permission storage.Rule
	if err := p.db.Where("type = ? and V1 = ? and V2 = ?", constant.TypePolicy, resource, action).First(&permission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, fmt.Errorf("not found permission")
		} else {
			return false, fmt.Errorf("other error: %w", err)
		}
	}

	var userRole storage.Rule
	if err := p.db.Where("type = ? and V0 = ? and V1 = ?").First(&userRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, fmt.Errorf("have no permission") 
		} else {
			return false, fmt.Errorf("other error: %w", err)
		}
	}

	return true, nil
}

func (p *Permission) AddPolicy(role, resource, action string) error {
	return p.db.Where("type = ? and V0 = ? and V1 = ? and V2 = ?",constant.TypePolicy, role, resource, action).FirstOrCreate(&storage.Rule{
		Type: constant.TypePolicy,
		V0: role,
		V1: resource,
		V2: action,
	}).Error
}

func (p *Permission) RemovePolicy(role, resource, action string) error {
	return p.db.Where("type = ? and V0 = ? and V1 = ? and V2 = ?", constant.TypePolicy, role, resource, action).Delete(&storage.Rule{}).Error
}

func (p *Permission) GetPolicy(role string) ([]storage.Rule, error) {
	var rules []storage.Rule
	if err := p.db.Where("type = ? and V0 = ?", constant.TypePolicy, role).Find(&rules).Error; err != nil {
		return nil, err
	}

	return rules, nil
}

func (p *Permission) AddGroup(userID string, role string) error {
	return p.db.Where("type = ? and V0 = ? and V1 = ?").FirstOrCreate(&storage.Rule{
		Type: constant.TypeGroup,
		V0: userID,
		V1: role,
	}).Error
}

func (p *Permission) RemoveGroup(userID string, role string) error {
	return p.db.Where("type = ? and V0 = ? and V1 = ?").Delete(&storage.Rule{}).Error
}