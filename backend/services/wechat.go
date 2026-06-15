package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

	envVersion := "develop"
	if config.AppConfig != nil && config.AppConfig.WeChat.EnvVersion != "" {
		envVersion = config.AppConfig.WeChat.EnvVersion
	}

	body := map[string]interface{}{
		"scene": scene,
		"page":  page,
		"check_path": false,
		"env_version": envVersion,
		"width": 280,
	}

	bodyBytes, _ := json.Marshal(body)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(bodyBytes)))
	req.ContentLength = int64(len(bodyBytes))
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wxacode request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read wxacode response failed: %w", err)
	}

	if resp.StatusCode != 200 {
		errMsg := string(respBody)
		if len(respBody) > 0 && respBody[0] == '{' {
			var errResp struct {
				ErrCode int    `json:"errcode"`
				ErrMsg  string `json:"errmsg"`
			}
			if json.Unmarshal(respBody, &errResp) == nil && errResp.ErrCode != 0 {
				errMsg = fmt.Sprintf("errcode=%d errmsg=%s", errResp.ErrCode, errResp.ErrMsg)
			}
		}
		return nil, fmt.Errorf("wxacode request failed status=%d %s", resp.StatusCode, errMsg)
	}

	if len(respBody) == 0 {
		return nil, fmt.Errorf("wxacode returned empty image body")
	}
	return &WXACodeUnlimitedResponse{Buffer: respBody}, nil
}

// --- wxa/generatescheme (URL Scheme for opening mini-program from H5) ---

type schemeCacheEntry struct {
	SchemeURL string
	ExpiresAt time.Time
}

var (
	schemeCache   = make(map[string]*schemeCacheEntry) // key: "shopID:tableNo"
	schemeCacheMu sync.RWMutex
)

type generateSchemeRequest struct {
	JumpWxa  generateSchemeJumpWxa `json:"jump_wxa"`
	IsExpire bool                   `json:"is_expire"`
}

type generateSchemeJumpWxa struct {
	Path       string `json:"path"`
	Query      string `json:"query"`
	EnvVersion string `json:"env_version"`
}

type generateSchemeResponse struct {
	OpenLink string `json:"openlink"`
	ErrCode  int    `json:"errcode"`
	ErrMsg   string `json:"errmsg"`
}

// GenerateURLScheme generates a WeChat URL Scheme for opening the mini-program.
// shopID and tableNo are encoded as query parameters on the target page.
// Returns the scheme URL like "weixin://dl/business/?t=TOKEN".
// Schemes are cached per (shopID, tableNo) and refreshed before expiry.
func GenerateURLScheme(shopID, tableNo string) (string, error) {
	cacheKey := shopID + ":" + tableNo

	// Check cache first (valid if more than 7 days until expiry)
	schemeCacheMu.RLock()
	entry, ok := schemeCache[cacheKey]
	schemeCacheMu.RUnlock()
	if ok && entry != nil && time.Now().Before(entry.ExpiresAt.AddDate(0, 0, -7)) {
		return entry.SchemeURL, nil
	}

	token, err := GetAccessToken()
	if err != nil {
		return "", fmt.Errorf("get access token: %w", err)
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/generatescheme?access_token=%s", token)

	envVersion := "release"
	if config.AppConfig != nil && config.AppConfig.WeChat.EnvVersion != "" {
		envVersion = config.AppConfig.WeChat.EnvVersion
	}

	body := generateSchemeRequest{
		JumpWxa: generateSchemeJumpWxa{
			Path:       "/pages/home/index",
			Query:      "shop_id=" + shopID + "&table_no=" + tableNo,
			EnvVersion: envVersion,
		},
		IsExpire: false, // permanent scheme
	}

	bodyBytes, _ := json.Marshal(body)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("scheme request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read scheme response failed: %w", err)
	}

	var result generateSchemeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse scheme response failed: %w", err)
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("wechat scheme error %d: %s", result.ErrCode, result.ErrMsg)
	}

	if result.OpenLink == "" {
		return "", fmt.Errorf("scheme response missing openlink")
	}

	// Cache the scheme (permanent schemes don't have an expiry, but we set a 30-day refresh cycle)
	schemeCacheMu.Lock()
	schemeCache[cacheKey] = &schemeCacheEntry{
		SchemeURL: result.OpenLink,
		ExpiresAt: time.Now().AddDate(0, 0, 30),
	}
	schemeCacheMu.Unlock()

	return result.OpenLink, nil
}