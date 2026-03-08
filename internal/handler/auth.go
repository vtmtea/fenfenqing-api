package handler

import (
	"bytes"
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

// GenerateQRCodeRequest 生成二维码请求
type GenerateQRCodeRequest struct {
	RoomID string `json:"roomId" binding:"required"`
}

// GenerateQRCode 生成小程序码
func (h *AuthHandler) GenerateQRCode(c *gin.Context) {
	var req GenerateQRCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 检查微信配置
	if h.wechat.AppID == "" || h.wechat.AppSecret == "" {
		response.InternalError(c, "服务器配置错误")
		return
	}

	// 获取 access token
	accessToken, err := h.getAccessToken()
	if err != nil {
		response.InternalError(c, "获取 access token 失败")
		return
	}

	// 调用微信接口生成小程序码
	qrCodeURL := fmt.Sprintf(
		"https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=%s",
		accessToken,
	)

	// 请求体
	reqBody := map[string]interface{}{
		"scene":       req.RoomID, // 场景值为房间号（最长 32 字符）
		"page":        "pages/join/join",
		"width":       430,
		"auto_color":  false,
		"is_hyaline":  false,
		"env_version": "trial", // release:正式版，trial:体验版，develop:开发版
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		response.InternalError(c, "参数序列化失败")
		return
	}

	resp, err := http.Post(qrCodeURL, "application/json", bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		response.InternalError(c, "生成二维码失败")
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.InternalError(c, "读取响应失败")
		return
	}

	// 检查是否返回错误（微信返回 JSON 表示错误）
	var errResp map[string]interface{}
	if err := json.Unmarshal(body, &errResp); err == nil {
		if errMsg, ok := errResp["errmsg"]; ok {
			response.InternalError(c, fmt.Sprintf("微信接口错误：%v", errMsg))
			return
		}
	}

	// 返回二维码图片的 base64
	c.Header("Content-Type", "image/png")
	c.Data(200, "image/png", body)
}

// getAccessToken 获取小程序全局 access token
func (h *AuthHandler) getAccessToken() (string, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		h.wechat.AppID,
		h.wechat.AppSecret,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if accessToken, ok := result["access_token"].(string); ok {
		return accessToken, nil
	}

	if errMsg, ok := result["errmsg"].(string); ok {
		return "", fmt.Errorf("获取 access token 失败：%s", errMsg)
	}

	return "", fmt.Errorf("获取 access token 失败")
}
