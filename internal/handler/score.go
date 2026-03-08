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
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的房间 ID")
		return
	}

	var scores []model.Score
	if err := h.db.Where("room_id = ?", id).Order("create_time DESC").Limit(50).Find(&scores).Error; err != nil {
		response.InternalError(c, "获取分数记录失败")
		return
	}

	response.Success(c, scores)
}

// AddScore 添加分数记录
func (h *ScoreHandler) AddScore(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c)
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的房间 ID")
		return
	}

	var req struct {
		Details    []model.ScoreDetail `json:"details" binding:"required,min=1"`
		Operator   string              `json:"operator"` // 操作者名称（前端传入）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "分数详情不能为空")
		return
	}

	// 检查房间是否存在
	var room model.Room
	if err := h.db.First(&room, id).Error; err != nil {
		response.NotFound(c)
		return
	}

	// 获取操作者信息
	var user model.User
	operatorName := req.Operator
	if err := h.db.First(&user, userID).Error; err == nil && req.Operator == "" {
		operatorName = user.Nickname
	}

	score := &model.Score{
		RoomID:     uint(id),
		OperatorID: userID.(uint),
		Operator:   operatorName,
		Details:    req.Details,
	}

	if err := h.db.Create(score).Error; err != nil {
		response.InternalError(c, "添加分数记录失败")
		return
	}

	response.Success(c, score)
}

// DeleteScore 删除分数记录
func (h *ScoreHandler) DeleteScore(c *gin.Context) {
	idStr := c.Param("id")
	scoreIDStr := c.Param("scoreID")

	id, err := strconv.ParseUint(idStr, 10, 32)
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
	if err := h.db.Where("id = ? AND room_id = ?", scoreID, id).First(&score).Error; err != nil {
		response.NotFound(c)
		return
	}

	if err := h.db.Delete(&score).Error; err != nil {
		response.InternalError(c, "删除分数记录失败")
		return
	}

	response.Success(c, nil)
}
