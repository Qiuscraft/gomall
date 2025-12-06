// Copyright 2024 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	"time"

	"github.com/cloudwego/biz-demo/gomall/app/payment/biz/dal/mysql"
	"github.com/cloudwego/biz-demo/gomall/app/payment/biz/model"
	"github.com/cloudwego/biz-demo/gomall/app/payment/infra/zhifufm"
	payment "github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/payment"
	"github.com/cloudwego/kitex/pkg/kerrors"
)

type ChargeService struct {
	ctx context.Context
} // NewChargeService new ChargeService
func NewChargeService(ctx context.Context) *ChargeService {
	return &ChargeService{ctx: ctx}
}

// Run create note info
func (s *ChargeService) Run(req *payment.ChargeReq) (resp *payment.ChargeResp, err error) {
	client := zhifufm.NewClientFromEnv()
	orderResp, err := client.StartOrder(s.ctx, zhifufm.StartOrderReq{
		AmountCents: req.Amount,
		OrderNo:     req.OrderId,
	})
	if err != nil {
		return nil, kerrors.NewBizStatusError(500, err.Error())
	}

	transactionID := orderResp.TransactionID
	if transactionID == "" {
		transactionID = req.OrderId
	}

	err = model.CreatePaymentLog(mysql.DB, s.ctx, &model.PaymentLog{
		UserId:        req.UserId,
		OrderId:       req.OrderId,
		TransactionId: transactionID,
		ProviderOrder: orderResp.TransactionID,
		PayURL:        orderResp.PayURL,
		Status:        "pending",
		Amount:        req.Amount,
		PayAt:         time.Now(),
	})
	if err != nil {
		return nil, err
	}

	return &payment.ChargeResp{TransactionId: transactionID, PayUrl: orderResp.PayURL}, nil
}
