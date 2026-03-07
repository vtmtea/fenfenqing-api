package model

import (
	"time"

	"gorm.io/gorm"
)

// Room 房间
type Room struct {
	ID          uint      `gorm:"primarykey" json:"_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	RoomID      string    `gorm:"size:10;uniqueIndex;not null" json:"roomId"`
	MemberCount int       `gorm:"default:0" json:"memberCount"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"autoUpdateTime" json:"updateTime"`
}

// Member 成员
type Member struct {
	ID         uint      `gorm:"primarykey" json:"_id"`
	RoomID     uint      `gorm:"index;not null" json:"roomId"`
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
	return db.AutoMigrate(&Room{}, &Member{}, &Score{})
}
