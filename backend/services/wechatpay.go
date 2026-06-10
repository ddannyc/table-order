package services

import (
	"context"
	"fmt"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/example/table-order/config"
)

// PrepayResult holds the parameters needed by wx.requestPayment on the frontend.
type PrepayResult struct {
	PrepayID  string `json:"prepay_id"`
	TimeStamp string `json:"time_stamp"`
	NonceStr  string `json:"nonce_str"`
	Package   string `json:"package"`
	SignType  string `json:"sign_type"`
	PaySign   string `json:"pay_sign"`
}

// CreateJSAPIPrepay calls WeChat Pay JSAPI unified order and returns
// the prepay parameters for the frontend to invoke wx.requestPayment.
// Returns empty PrepayResult with no error when WeChat Pay is not configured.
func CreateJSAPIPrepay(ctx context.Context, openID, orderNo, description string, amountCents int64) (*PrepayResult, error) {
	if config.WxPayClient == nil {
		return nil, fmt.Errorf("wechat pay not configured")
	}

	cfg := config.AppConfig.WeChat
	svc := jsapi.JsapiApiService{Client: config.WxPayClient}

	resp, _, err := svc.PrepayWithRequestPayment(ctx,
		jsapi.PrepayRequest{
			Appid:       core.String(cfg.AppID),
			Mchid:       core.String(cfg.MchID),
			Description: core.String(description),
			OutTradeNo:  core.String(orderNo),
			NotifyUrl:   core.String(cfg.PayNotifyURL),
			Amount: &jsapi.Amount{
				Total: core.Int64(amountCents),
			},
			Payer: &jsapi.Payer{
				Openid: core.String(openID),
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("wechat prepay failed: %w", err)
	}

	return &PrepayResult{
		PrepayID:  *resp.PrepayId,
		TimeStamp: *resp.TimeStamp,
		NonceStr:  *resp.NonceStr,
		Package:   *resp.Package,
		SignType:  *resp.SignType,
		PaySign:   *resp.PaySign,
	}, nil
}
