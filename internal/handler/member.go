package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vtmtea/fenfenqing-api/internal/model"
	"github.com/vtmtea/fenfenqing-api/pkg/response"
)

// MemberHandler 成员处理器
type MemberHandler struct {
	db *model.DB
}

// NewMemberHandler 创建成员处理器
func NewMemberHandler(db *model.DB) *MemberHandler {
	return &MemberHandler{db: db}
}

// GetMemberList 获取成员列表
func (h *MemberHandler) GetMemberList(c *gin.Context) {
	roomIDStr := c.Param("roomID")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的房间 ID")
		return
	}

	var members []model.Member
	if err := h.db.Where("room_id = ?", roomID).Find(&members).Error; err != nil {
		response.InternalError(c, "获取成员列表失败")
		return
	}

	response.Success(c, members)
}

// AddMember 添加成员
func (h *MemberHandler) AddMember(c *gin.Context) {
	roomIDStr := c.Param("roomID")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的房间 ID")
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "成员名称不能为空")
		return
	}

	// 检查房间是否存在
	var room model.Room
	if err := h.db.First(&room, roomID).Error; err != nil {
		response.NotFound(c)
		return
	}

	member := &model.Member{
		RoomID: uint(roomID),
		Name:   req.Name,
	}

	if err := h.db.Create(member).Error; err != nil {
		response.InternalError(c, "添加成员失败")
		return
	}

	// 更新房间成员数量
	h.db.Model(&room).Update("member_count", room.MemberCount+1)

	response.Success(c, member)
}

// DeleteMember 删除成员
func (h *MemberHandler) DeleteMember(c *gin.Context) {
	roomIDStr := c.Param("roomID")
	memberIDStr := c.Param("memberID")

	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的房间 ID")
		return
	}

	memberID, err := strconv.ParseUint(memberIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的成员 ID")
		return
	}

	// 检查房间是否存在
	var room model.Room
	if err := h.db.First(&room, roomID).Error; err != nil {
		response.NotFound(c)
		return
	}

	// 删除成员
	var member model.Member
	if err := h.db.Where("id = ? AND room_id = ?", memberID, roomID).First(&member).Error; err != nil {
		response.NotFound(c)
		return
	}

	if err := h.db.Delete(&member).Error; err != nil {
		response.InternalError(c, "删除成员失败")
		return
	}

	// 更新房间成员数量
	if room.MemberCount > 0 {
		h.db.Model(&room).Update("member_count", room.MemberCount-1)
	}

	response.Success(c, nil)
}
