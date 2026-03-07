package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vtmtea/fenfenqing-api/internal/model"
	"github.com/vtmtea/fenfenqing-api/pkg/response"
	"gorm.io/gorm"
)

// ScoreHandler 分数处理器
type ScoreHandler struct {
	db *gorm.DB
}

// NewScoreHandler 创建分数处理器
func NewScoreHandler(db *gorm.DB) *ScoreHandler {
	return &ScoreHandler{db: db}
}

// GetScoreList 获取分数记录列表
func (h *ScoreHandler) GetScoreList(c *gin.Context) {
	roomIDStr := c.Param("roomID")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的房间 ID")
		return
	}

	var scores []model.Score
	if err := h.db.Where("room_id = ?", roomID).Order("create_time DESC").Limit(50).Find(&scores).Error; err != nil {
		response.InternalError(c, "获取分数记录失败")
		return
	}

	response.Success(c, scores)
}

// AddScore 添加分数记录
func (h *ScoreHandler) AddScore(c *gin.Context) {
	roomIDStr := c.Param("roomID")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的房间 ID")
		return
	}

	var req struct {
		Details []model.ScoreDetail `json:"details" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "分数详情不能为空")
		return
	}

	// 检查房间是否存在
	var room model.Room
	if err := h.db.First(&room, roomID).Error; err != nil {
		response.NotFound(c)
		return
	}

	score := &model.Score{
		RoomID:  uint(roomID),
		Details: req.Details,
	}

	if err := h.db.Create(score).Error; err != nil {
		response.InternalError(c, "添加分数记录失败")
		return
	}

	response.Success(c, score)
}

// DeleteScore 删除分数记录
func (h *ScoreHandler) DeleteScore(c *gin.Context) {
	roomIDStr := c.Param("roomID")
	scoreIDStr := c.Param("scoreID")

	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的房间 ID")
		return
	}

	scoreID, err := strconv.ParseUint(scoreIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的分数记录 ID")
		return
	}

	var score model.Score
	if err := h.db.Where("id = ? AND room_id = ?", scoreID, roomID).First(&score).Error; err != nil {
		response.NotFound(c)
		return
	}

	if err := h.db.Delete(&score).Error; err != nil {
		response.InternalError(c, "删除分数记录失败")
		return
	}

	response.Success(c, nil)
}
