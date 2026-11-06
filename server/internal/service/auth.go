package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"xmf/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 登录相关业务
type AuthService struct {
	DB         *gorm.DB
	WxAppID    string
	WxSecret   string
	HTTPClient *http.Client
}

func NewAuthService(db *gorm.DB, appID, secret string) *AuthService {
	return &AuthService{DB: db, WxAppID: appID, WxSecret: secret, HTTPClient: &http.Client{Timeout: 10 * time.Second}}
}

// DevLogin 本地开发模拟登录：openid 即用户，不存在则创建。仅 XMF_DEV_LOGIN=1 时由 api 层放行。
func (s *AuthService) DevLogin(openID string) (*model.User, error) {
	if openID == "" {
		return nil, errors.New("openid 不能为空")
	}
	return s.findOrCreateUser(openID)
}

type wxSessionResp struct {
	OpenID  string `json:"openid"`
	ErrMsg  string `json:"errmsg"`
	ErrCode int    `json:"errcode"`
}

// WxLogin 微信 code 换 openid
func (s *AuthService) WxLogin(code string) (*model.User, error) {
	if s.WxAppID == "" || s.WxSecret == "" {
		return nil, errors.New("小程序未配置登录参数，请使用开发者模拟登录")
	}
	if code == "" {
		return nil, errors.New("code 不能为空")
	}
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.WxAppID, s.WxSecret, code)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.New("微信登录服务暂时不可用，请稍后再试")
	}
	defer resp.Body.Close()
	var sr wxSessionResp
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil || sr.OpenID == "" {
		return nil, errors.New("微信登录失败，请稍后再试")
	}
	return s.findOrCreateUser(sr.OpenID)
}

func (s *AuthService) findOrCreateUser(openID string) (*model.User, error) {
	var u model.User
	err := s.DB.Where("open_id = ?", openID).First(&u).Error
	if err == nil {
		if u.Status != 1 {
			return nil, errors.New("账号已被禁用")
		}
		return &u, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	u = model.User{OpenID: openID, Status: 1}
	if err := s.DB.Create(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// AdminLogin 管理员登录
func (s *AuthService) AdminLogin(username, password string) (*model.Admin, error) {
	var a model.Admin
	if err := s.DB.Where("username = ?", username).First(&a).Error; err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if a.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}
	if bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(password)) != nil {
		return nil, errors.New("用户名或密码错误")
	}
	return &a, nil
}
