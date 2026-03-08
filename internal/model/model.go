package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户
type User struct {
	ID           uint      `gorm:"primarykey" json:"_id"`
	OpenID       string    `gorm:"size:64;uniqueIndex;not null" json:"openid"`
	SessionKey   string    `gorm:"size:64" json:"-"`
	Nickname     string    `gorm:"size:100" json:"nickname"`
	AvatarURL    string    `gorm:"size:500" json:"avatarUrl"`
	Phone        string    `gorm:"size:20" json:"phone"`
	LastLoginAt  time.Time `json:"lastLoginAt"`
	CreateTime   time.Time `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime   time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

// Room 房间
type Room struct {
	ID          uint      `gorm:"primarykey" json:"_id"`
	UserID      uint      `gorm:"index;not null" json:"userId"` // 创建者 ID
	Name        string    `gorm:"size:100;not null" json:"name"`
	RoomID      string    `gorm:"size:10;uniqueIndex;not null" json:"roomId"`
	MemberCount int       `gorm:"default:0" json:"memberCount"`
	Status      int       `gorm:"default:0;comment:0-进行中，1-已关闭" json:"status"`
	ClosedAt    *time.Time `gorm:"index" json:"closedAt,omitempty"`
	ClosedBy    uint      `gorm:"default:0" json:"closedBy"` // 关闭者 ID
	CreateTime  time.Time `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

// RoomMember 房间成员关联
type RoomMember struct {
	ID         uint      `gorm:"primarykey" json:"_id"`
	RoomID     uint      `gorm:"index;not null" json:"roomId"`
	UserID     uint      `gorm:"index;not null" json:"userId"`
	Name       string    `gorm:"size:50;not null" json:"name"`
	CreateTime time.Time `gorm:"autoCreateTime" json:"createTime"`
}

// Score 分数记录
type Score struct {
	ID         uint           `gorm:"primarykey" json:"_id"`
	RoomID     uint           `gorm:"index;not null" json:"roomId"`
	Details    ScoreDetails   `gorm:"serializer:json" json:"details"`
	CreateTime time.Time      `gorm:"autoCreateTime" json:"createTime"`
}

// ScoreDetail 分数详情
type ScoreDetail struct {
	MemberID uint   `json:"memberId"`
	Name     string `json:"name"`
	Value    int    `json:"value"`
}

type ScoreDetails []ScoreDetail

// InitDB 初始化数据库表
func InitDB(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Room{}, &RoomMember{}, &Score{})
}
