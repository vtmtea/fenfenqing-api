package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vtmtea/fenfenqing-api/internal/model"
	"github.com/vtmtea/fenfenqing-api/internal/websocket"
	"github.com/vtmtea/fenfenqing-api/pkg/response"
	"gorm.io/gorm"
)

// 全局 Hub 引用（由 main.go 初始化时设置）
var Hub *websocket.Hub

// SetHub 设置全局 Hub
func SetHub(hub *websocket.Hub) {
	Hub = hub
}

// MemberHandler 成员处理器
type MemberHandler struct {
	db *gorm.DB
}

// NewMemberHandler 创建成员处理器
func NewMemberHandler(db *gorm.DB) *MemberHandler {
	return &MemberHandler{db: db}
}

// GetMemberList 获取成员列表
func (h *MemberHandler) GetMemberList(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的房间 ID")
		return
	}

	var members []model.RoomMember
	if err := h.db.Where("room_id = ?", roomID).Find(&members).Error; err != nil {
		response.InternalError(c, "获取成员列表失败")
		return
	}

	// 关联查询用户信息，获取头像
	type MemberWithAvatar struct {
		model.RoomMember
		AvatarURL string `json:"avatarUrl"`
	}
	result := make([]MemberWithAvatar, 0, len(members))
	for _, member := range members {
		var user model.User
		if err := h.db.First(&user, member.UserID).Error; err == nil {
			result = append(result, MemberWithAvatar{
				RoomMember: member,
				AvatarURL:  user.AvatarURL,
			})
		} else {
			result = append(result, MemberWithAvatar{
				RoomMember: member,
				AvatarURL:  "",
			})
		}
	}

	response.Success(c, result)
}

// AddMember 添加成员
func (h *MemberHandler) AddMember(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c)
		return
	}

	openID := c.GetString("openid")
	fmt.Printf("=== AddMember 调试信息 ===\n")
	fmt.Printf("userID: %d, openID: %s\n", userID, openID)

	roomIDStr := c.Param("id")
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

	// 检查是否已是成员
	var existing model.RoomMember
	if err := h.db.Where("room_id = ? AND user_id = ?", roomID, userID).First(&existing).Error; err == nil {
		// 已是成员，更新名称
		fmt.Printf("成员已存在，更新名称\n")
		h.db.Model(&existing).Update("name", req.Name)
		response.Success(c, existing)
		return
	}

	member := &model.RoomMember{
		RoomID: uint(roomID),
		UserID: userID.(uint),
		Name:   req.Name,
	}

	fmt.Printf("创建新成员：%+v\n", member)
	if err := h.db.Create(member).Error; err != nil {
		response.InternalError(c, "添加成员失败")
		return
	}

	// 更新房间成员数量
	h.db.Model(&room).Update("member_count", room.MemberCount+1)

	// 查询用户信息获取头像
	var user model.User
	avatarURL := ""
	if err := h.db.First(&user, userID).Error; err == nil {
		avatarURL = user.AvatarURL
	}

	// 广播新成员加入消息
	if Hub != nil {
		Hub.BroadcastToRoom(uint(roomID), &websocket.Message{
			Type: "member_join",
			Data: gin.H{
				"_id":       member.ID,
				"userId":    member.UserID,
				"name":      member.Name,
				"avatarUrl": avatarURL,
			},
		})
	}

	response.Success(c, member)
}

// DeleteMember 删除成员
func (h *MemberHandler) DeleteMember(c *gin.Context) {
	roomIDStr := c.Param("id")
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
	var member model.RoomMember
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
