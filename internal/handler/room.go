package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vtmtea/fenfenqing-api/internal/model"
	"github.com/vtmtea/fenfenqing-api/pkg/response"
	"gorm.io/gorm"
)

// RoomHandler 房间处理器
type RoomHandler struct {
	db *gorm.DB
}

// NewRoomHandler 创建房间处理器
func NewRoomHandler(db *gorm.DB) *RoomHandler {
	return &RoomHandler{db: db}
}

// CreateRoom 创建房间
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "房间名称不能为空")
		return
	}

	// 生成唯一房间号
	roomID := model.GenerateRoomID()
	for h.db.Where("room_id = ?", roomID).First(&model.Room{}).Error == nil {
		roomID = model.GenerateRoomID()
	}

	room := &model.Room{
		Name:   req.Name,
		RoomID: roomID,
	}

	if err := h.db.Create(room).Error; err != nil {
		response.InternalError(c, "创建房间失败")
		return
	}

	response.Success(c, room)
}

// GetRoomList 获取房间列表
func (h *RoomHandler) GetRoomList(c *gin.Context) {
	var rooms []model.Room

	if err := h.db.Order("update_time DESC").Find(&rooms).Error; err != nil {
		response.InternalError(c, "获取房间列表失败")
		return
	}

	response.Success(c, rooms)
}

// GetRoomByID 根据 ID 获取房间
func (h *RoomHandler) GetRoomByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的房间 ID")
		return
	}

	var room model.Room
	if err := h.db.First(&room, id).Error; err != nil {
		response.NotFound(c)
		return
	}

	response.Success(c, room)
}

// GetRoomByRoomID 根据房间号获取房间
func (h *RoomHandler) GetRoomByRoomID(c *gin.Context) {
	roomID := c.Param("roomID")

	var room model.Room
	if err := h.db.Where("room_id = ?", roomID).First(&room).Error; err != nil {
		response.NotFound(c)
		return
	}

	response.Success(c, room)
}

// DeleteRoom 删除房间
func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的房间 ID")
		return
	}

	// 使用事务删除房间及相关数据
	tx := h.db.Begin()
	if err := tx.Delete(&model.Room{}, id).Error; err != nil {
		tx.Rollback()
		response.InternalError(c, "删除房间失败")
		return
	}

	// 删除成员
	tx.Where("room_id = ?", id).Delete(&model.Member{})
	// 删除分数记录
	tx.Where("room_id = ?", id).Delete(&model.Score{})

	tx.Commit()

	response.Success(c, nil)
}
