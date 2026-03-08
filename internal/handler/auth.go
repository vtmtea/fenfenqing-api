package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vtmtea/fenfenqing-api/internal/model"
	"github.com/vtmtea/fenfenqing-api/pkg/jwt"
	"github.com/vtmtea/fenfenqing-api/pkg/response"
	"gorm.io/gorm"
)

// WeChatConfig 微信配置
type WeChatConfig struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

// AuthHandler 认证处理器
type AuthHandler struct {
	db     *gorm.DB
	wechat WeChatConfig
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(db *gorm.DB, appID, appSecret string) *AuthHandler {
	return &AuthHandler{
		db: db,
		wechat: WeChatConfig{
			AppID:     appID,
			AppSecret: appSecret,
		},
	}
}

// Code2SessionResponse 微信 code2Session 响应
type Code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid,omitempty"`
	ErrCode    int    `json:"errcode,omitempty"`
	ErrMsg     string `json:"errmsg,omitempty"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Code      string `json:"code" binding:"required"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
	Phone     string `json:"phone"`
}

// Login 微信登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 检查微信配置
	if h.wechat.AppID == "" || h.wechat.AppSecret == "" {
		response.InternalError(c, "服务器配置错误")
		return
	}

	// 调用微信接口获取 openid
	openID, sessionKey, err := h.code2Session(req.Code)
	if err != nil {
		response.InternalError(c, "微信登录失败")
		return
	}

	if openID == "" {
		response.BadRequest(c, "获取用户信息失败")
		return
	}

	// 查找或创建用户
	var user model.User
	result := h.db.Where("open_id = ?", openID).First(&user)

	now := time.Now()
	if result.Error == gorm.ErrRecordNotFound {
		// 创建新用户
		user = model.User{
			OpenID:      openID,
			SessionKey:  sessionKey,
			Nickname:    req.Nickname,
			AvatarURL:   req.AvatarURL,
			Phone:       req.Phone,
			LastLoginAt: now,
		}
		if err := h.db.Create(&user).Error; err != nil {
			response.InternalError(c, "创建用户失败")
			return
		}
	} else if result.Error == nil {
		// 更新用户信息
		updates := map[string]interface{}{
			"session_key":  sessionKey,
			"last_login_at": now,
		}
		if req.Nickname != "" {
			updates["nickname"] = req.Nickname
		}
		if req.AvatarURL != "" {
			updates["avatar_url"] = req.AvatarURL
		}
		if req.Phone != "" {
			updates["phone"] = req.Phone
		}
		h.db.Model(&user).Updates(updates)
	} else {
		response.InternalError(c, "查询用户失败")
		return
	}

	// 生成 JWT token
	token, err := jwt.GenerateToken(user.ID, user.OpenID)
	if err != nil {
		response.InternalError(c, "生成 token 失败")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"_id":       user.ID,
			"openid":    user.OpenID,
			"nickname":  user.Nickname,
			"avatarUrl": user.AvatarURL,
			"phone":     user.Phone,
		},
	})
}

// GetUserInfo 获取用户信息
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c)
		return
	}

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.NotFound(c)
		return
	}

	response.Success(c, gin.H{
		"_id":       user.ID,
		"openid":    user.OpenID,
		"nickname":  user.Nickname,
		"avatarUrl": user.AvatarURL,
		"phone":     user.Phone,
	})
}

// UpdateUserInfo 更新用户信息
func (h *AuthHandler) UpdateUserInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c)
		return
	}

	var req struct {
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarUrl"`
		Phone     string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.NotFound(c)
		return
	}

	updates := map[string]interface{}{}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.AvatarURL != "" {
		updates["avatar_url"] = req.AvatarURL
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}

	if len(updates) > 0 {
		h.db.Model(&user).Updates(updates)
	}

	response.Success(c, gin.H{
		"_id":       user.ID,
		"openid":    user.OpenID,
		"nickname":  user.Nickname,
		"avatarUrl": user.AvatarURL,
		"phone":     user.Phone,
	})
}

// code2Session 调用微信接口获取 openid
func (h *AuthHandler) code2Session(code string) (string, string, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		h.wechat.AppID,
		h.wechat.AppSecret,
		code,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var result Code2SessionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", err
	}

	if result.ErrCode != 0 {
		return "", "", fmt.Errorf("微信接口错误：%s", result.ErrMsg)
	}

	return result.OpenID, result.SessionKey, nil
}
