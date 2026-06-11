package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/example/table-order/services"
)

const scanHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
  <title>正在打开点餐小程序...</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Helvetica Neue", sans-serif;
      background: #f5f5f5;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      min-height: 100vh;
      padding: 40rpx 20rpx;
      text-align: center;
    }
    .container {
      background: #fff;
      border-radius: 16px;
      padding: 48px 32px;
      max-width: 360px;
      width: 100%%;
      box-shadow: 0 2px 12px rgba(0,0,0,0.08);
    }
    .icon {
      width: 72px;
      height: 72px;
      background: #07c160;
      border-radius: 50%%;
      display: flex;
      align-items: center;
      justify-content: center;
      margin: 0 auto 24px;
      font-size: 36px;
      color: #fff;
    }
    .title {
      font-size: 20px;
      font-weight: 600;
      color: #1a1a1a;
      margin-bottom: 12px;
    }
    .desc {
      font-size: 14px;
      color: #999;
      line-height: 1.5;
    }
    .btn {
      display: inline-block;
      margin-top: 32px;
      padding: 12px 32px;
      background: #07c160;
      color: #fff;
      border-radius: 24px;
      font-size: 16px;
      text-decoration: none;
    }
    .error-icon {
      background: #f5f5f5;
      color: #999;
    }
    .manual {
      margin-top: 24px;
      padding-top: 24px;
      border-top: 1px solid #eee;
      font-size: 13px;
      color: #bbb;
    }
    .hide { display: none; }
  </style>
</head>
<body>
  <div class="container">
    <div class="icon" id="statusIcon">&#x2714;</div>
    <div class="title" id="statusTitle">正在打开点餐小程序...</div>
    <div class="desc" id="statusDesc">请稍候，即将跳转到点餐页面</div>
    <div class="manual hide" id="manualBox">
      <p>如果未自动打开，请点击下方按钮</p>
      <a class="btn" href="%s">打开小程序</a>
    </div>
  </div>
  <script>
    setTimeout(function() {
      window.location.href = "%s";
    }, 100);
    setTimeout(function() {
      document.getElementById('manualBox').classList.remove('hide');
    }, 3000);
  </script>
</body>
</html>`

const scanErrorHTML = `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
  <title>打开失败</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Helvetica Neue", sans-serif;
      background: #f5f5f5;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      min-height: 100vh;
      padding: 40rpx 20rpx;
      text-align: center;
    }
    .container {
      background: #fff;
      border-radius: 16px;
      padding: 48px 32px;
      max-width: 360px;
      width: 100%%;
      box-shadow: 0 2px 12px rgba(0,0,0,0.08);
    }
    .icon {
      width: 72px;
      height: 72px;
      background: #f5f5f5;
      border-radius: 50%%;
      display: flex;
      align-items: center;
      justify-content: center;
      margin: 0 auto 24px;
      font-size: 36px;
      color: #999;
    }
    .title {
      font-size: 20px;
      font-weight: 600;
      color: #1a1a1a;
      margin-bottom: 12px;
    }
    .desc {
      font-size: 14px;
      color: #999;
      line-height: 1.5;
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="icon">&#x26A0;</div>
    <div class="title">%s</div>
    <div class="desc">%s</div>
  </div>
</body>
</html>`

// ScanRedirect handles GET /scan — the H5 redirect page for table QR codes.
// It generates a WeChat URL Scheme and serves an HTML page that auto-redirects.
func ScanRedirect(c *gin.Context) {
	shopID := c.Query("shop_id")
	tableNo := c.Query("table_no")
	token := c.Query("token")

	if shopID == "" || tableNo == "" {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusBadRequest, fmt.Sprintf(scanErrorHTML,
			"参数错误", "二维码缺少必要参数，请重新扫描或联系店员"))
		return
	}

	schemeURL, err := services.GenerateURLScheme(shopID, tableNo)
	if err != nil {
		// Log the error but serve a friendly page
		fmt.Printf("[scan] GenerateURLScheme error: %v (shop_id=%s, table_no=%s, token=%s)\n",
			err, shopID, tableNo, token)

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, fmt.Sprintf(scanErrorHTML,
			"暂时无法打开小程序",
			"请手动打开微信小程序「点餐」并扫描桌上二维码"))
		return
	}

	escapedScheme := url.QueryEscape(schemeURL)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, fmt.Sprintf(scanHTMLTemplate, escapedScheme, escapedScheme))
}
