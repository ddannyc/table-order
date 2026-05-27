package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/example/table-order/config"
)

// --- jscode2session ---

type WeChatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func GetWeChatSession(code string) (*WeChatSessionResponse, error) {
	if config.AppConfig == nil || config.AppConfig.WeChat.AppID == "" || config.AppConfig.WeChat.AppSecret == "" {
		return nil, fmt.Errorf("wechat appid/appsecret not configured")
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		config.AppConfig.WeChat.AppID,
		config.AppConfig.WeChat.AppSecret,
		code,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("wechat request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	var result WeChatSessionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error %d: %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// --- access_token (cached) ---

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

var (
	tokenCache   string
	tokenExpires time.Time
	tokenMu      sync.Mutex
)

func GetAccessToken() (string, error) {
	tokenMu.Lock()
	defer tokenMu.Unlock()

	if tokenCache != "" && time.Now().Before(tokenExpires) {
		return tokenCache, nil
	}

	if config.AppConfig == nil || config.AppConfig.WeChat.AppID == "" || config.AppConfig.WeChat.AppSecret == "" {
		return "", fmt.Errorf("wechat appid/appsecret not configured")
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		config.AppConfig.WeChat.AppID,
		config.AppConfig.WeChat.AppSecret,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response failed: %w", err)
	}

	var result tokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse token response failed: %w", err)
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("wechat token error %d: %s", result.ErrCode, result.ErrMsg)
	}

	tokenCache = result.AccessToken
	tokenExpires = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)
	return tokenCache, nil
}

// --- wxacode.getUnlimited ---

type WXACodeUnlimitedResponse struct {
	Buffer  []byte
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// GetWXACodeUnlimited calls wxacode.getUnlimited to generate a mini-program QR code.
// scene: max 32 chars, visible in App.onLaunch/onShow scene param.
// page: must already be published (not draft) or be the mini-program's main page.
func GetWXACodeUnlimited(scene, page string) (*WXACodeUnlimitedResponse, error) {
	token, err := GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=%s", token)

	body := map[string]interface{}{
		"scene": scene,
		"page":  page,
		"check_path": false,
		"env_version": "release",
		"width": 280,
	}

	bodyBytes, _ := json.Marshal(body)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(url, "application/json", io.NopCloser(bytes.NewReader(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("wxacode request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read wxacode response failed: %w", err)
	}

	// On success, the response is raw image bytes (not JSON).
	// On error, it returns JSON with errcode and errmsg.
	contentType := resp.Header.Get("Content-Type")
	fmt.Printf("[wechat] wxacode response status=%d content-type=%s body_len=%d body=%s\n", resp.StatusCode, contentType, len(respBody), string(respBody))

	// Check for error JSON first (regardless of content-type)
	if len(respBody) > 0 && respBody[0] == '{' {
		var errResp struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.ErrCode != 0 {
			return nil, fmt.Errorf("wxacode error %d: %s", errResp.ErrCode, errResp.ErrMsg)
		}
	}

	// Non-200 status with no error JSON
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("wxacode request failed status=%d body=%s", resp.StatusCode, string(respBody))
	}

	// Success — raw image
	if len(respBody) == 0 {
		return nil, fmt.Errorf("wxacode returned empty image body")
	}
	return &WXACodeUnlimitedResponse{Buffer: respBody}, nil
}