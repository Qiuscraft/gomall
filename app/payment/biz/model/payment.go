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

package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type PaymentLog struct {
	Base
	UserId        uint32    `json:"user_id"`
	OrderId       string    `json:"order_id"`
	TransactionId string    `json:"transaction_id"`
	ProviderOrder string    `json:"provider_order"`
	PayURL        string    `json:"pay_url"`
	Status        string    `json:"status"`
	Amount        uint64    `json:"amount"`
	PayAt         time.Time `json:"pay_at"`
}

func (p PaymentLog) TableName() string {
	return "payment"
}

func CreatePaymentLog(db *gorm.DB, ctx context.Context, payment *PaymentLog) error {
	return db.WithContext(ctx).Model(&PaymentLog{}).Create(payment).Error
}

func GetPaymentLogByOrderID(db *gorm.DB, ctx context.Context, orderID string) (*PaymentLog, error) {
	var p PaymentLog
	err := db.WithContext(ctx).Model(&PaymentLog{}).Where("order_id = ?", orderID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func MarkPaymentPaid(db *gorm.DB, ctx context.Context, orderID, transactionID, providerOrder, payURL string, payAt time.Time) error {
	return db.WithContext(ctx).Model(&PaymentLog{}).
		Where("order_id = ?", orderID).
		Updates(map[string]any{
			"transaction_id": transactionID,
			"provider_order": providerOrder,
			"pay_url":        payURL,
			"status":         "paid",
			"pay_at":         payAt,
		}).Error
}
