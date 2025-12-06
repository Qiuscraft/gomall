package zhifufm

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Client 封装支付FM最小能力：下单和签名规则。
type Client struct {
	BaseURL     string
	MerchantNum string
	SecretKey   string
	NotifyURL   string
	ReturnURL   string
	PayType     string
	httpCli     *http.Client
}

// StartOrderReq 下单入参，AmountCents 为分单位。
type StartOrderReq struct {
	AmountCents uint64
	OrderNo     string
	NotifyURL   string
	ReturnURL   string
	PayType     string
}

// StartOrderResp 下单返回的核心字段。
type StartOrderResp struct {
	TransactionID string
	PayURL        string
}

// NewClientFromEnv 根据环境变量构造客户端。
func NewClientFromEnv() *Client {
	timeout := 5 * time.Second
	if v := os.Getenv("PAYMENT_HTTP_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			timeout = d
		}
	}
	return &Client{
		BaseURL:     getenvDefault("ZHIFUFM_BASE_URL", ""),
		MerchantNum: os.Getenv("ZHIFUFM_MERCHANT_NUM"),
		SecretKey:   os.Getenv("ZHIFUFM_SECRET_KEY"),
		NotifyURL:   os.Getenv("ZHIFUFM_NOTIFY_URL"),
		ReturnURL:   os.Getenv("ZHIFUFM_RETURN_URL"),
		PayType:     getenvDefault("ZHIFUFM_PAY_TYPE", "alipay"),
		httpCli:     &http.Client{Timeout: timeout},
	}
}

// StartOrder 调用支付FM创建订单。
func (c *Client) StartOrder(ctx context.Context, req StartOrderReq) (*StartOrderResp, error) {
	if c.MerchantNum == "" || c.SecretKey == "" || c.BaseURL == "" {
		return nil, errors.New("支付FM配置缺失，请检查环境变量")
	}

	notify := firstNonEmpty(req.NotifyURL, c.NotifyURL)
	if notify == "" {
		return nil, errors.New("缺少 notifyUrl")
	}

	payType := firstNonEmpty(req.PayType, c.PayType)
	amountStr := formatAmountYuan(req.AmountCents)

	form := url.Values{}
	form.Set("merchantNum", c.MerchantNum)
	form.Set("orderNo", req.OrderNo)
	form.Set("amount", amountStr)
	form.Set("notifyUrl", notify)
	form.Set("payType", payType)
	if ru := firstNonEmpty(req.ReturnURL, c.ReturnURL); ru != "" {
		form.Set("returnUrl", ru)
	}
	form.Set("returnType", "json")
	form.Set("sign", signStartOrder(c.MerchantNum, req.OrderNo, amountStr, notify, c.SecretKey))

	endpoint := strings.TrimRight(c.BaseURL, "/") + "/startOrder"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpCli.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("startOrder http status=%d body=%s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Success bool   `json:"success"`
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		Data    struct {
			ID     string `json:"id"`
			PayURL string `json:"payUrl"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("startOrder decode fail: %w", err)
	}
	if !parsed.Success || parsed.Code >= 300 {
		return nil, fmt.Errorf("startOrder fail code=%d msg=%s", parsed.Code, parsed.Msg)
	}

	txn := parsed.Data.ID
	if txn == "" {
		txn = req.OrderNo
	}

	return &StartOrderResp{TransactionID: txn, PayURL: parsed.Data.PayURL}, nil
}

// VerifySignature 校验支付FM回调签名。
func VerifySignature(state, merchantNum, orderNo, amount, secret, sign string) bool {
	raw := state + merchantNum + orderNo + amount + secret
	sum := md5.Sum([]byte(raw))
	return hex.EncodeToString(sum[:]) == strings.ToLower(sign)
}

func signStartOrder(merchantNum, orderNo, amount, notifyURL, secret string) string {
	raw := merchantNum + orderNo + amount + notifyURL + secret
	sum := md5.Sum([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func formatAmountYuan(cents uint64) string {
	// 下单金额字符串，保留2位小数。
	return fmt.Sprintf("%.2f", float64(cents)/100)
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
