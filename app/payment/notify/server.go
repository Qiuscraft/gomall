package notify

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/cloudwego/biz-demo/gomall/app/payment/biz/dal/mysql"
	"github.com/cloudwego/biz-demo/gomall/app/payment/biz/model"
	"github.com/cloudwego/biz-demo/gomall/app/payment/conf"
	"github.com/cloudwego/biz-demo/gomall/app/payment/infra/rpc"
	"github.com/cloudwego/biz-demo/gomall/app/payment/infra/zhifufm"
	"github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/order"
	"github.com/cloudwego/kitex/pkg/klog"
)

// Start 启动一个简单的 HTTP 回调服务器，用于接收支付FM异步通知。
func Start(ctx context.Context) error {
	notifyURL := conf.GetEnvNotifyURL()
	path := "/payment/notify"
	addr := ":8089"
	if notifyURL != "" {
		if u, err := url.Parse(notifyURL); err == nil {
			if u.Path != "" {
				path = u.Path
			}
		}
	}

	http.HandleFunc(path, handleNotify)
	klog.Infof("notify server listening on %s%s", addr, path)

	srv := &http.Server{Addr: addr}
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	return srv.ListenAndServe()
}

func handleNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	q := r.URL.Query()
	state := q.Get("state")
	merchantNum := q.Get("merchantNum")
	orderNo := q.Get("orderNo")
	amount := q.Get("amount")
	sign := q.Get("sign")
	platformOrder := q.Get("platformOrderNo")
	actualPay := q.Get("actualPayAmount")
	payTimeStr := q.Get("payTime")

	if !zhifufm.VerifySignature(state, merchantNum, orderNo, amount, conf.GetSecretKey(), sign) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid sign"))
		return
	}
	if state != "1" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("state not paid"))
		return
	}

	logItem, err := model.GetPaymentLogByOrderID(mysql.DB, r.Context(), orderNo)
	if err != nil {
		klog.Error("payment log not found", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	payAt := time.Now()
	if payTimeStr != "" {
		if parsed, err := time.Parse("2006-01-02 15:04:05", payTimeStr); err == nil {
			payAt = parsed
		}
	}

	// 回写支付日志，标记已支付。
	_ = model.MarkPaymentPaid(mysql.DB, r.Context(), orderNo, platformOrder, platformOrder, logItem.PayURL, payAt)

	// 通知订单服务支付完成。
	if rpc.OrderClient != nil {
		_, _ = rpc.OrderClient.MarkOrderPaid(r.Context(), &order.MarkOrderPaidReq{UserId: logItem.UserId, OrderId: orderNo})
	}

	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte("success"))
	klog.Infof("order %s paid, actual=%s, platform=%s", orderNo, actualPay, platformOrder)
}
