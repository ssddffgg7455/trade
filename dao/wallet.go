package dao

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Wallet struct {
	Id           int             `gorm:"primaryKey;type:int;index;not null" json:"id"`      // 钱包id
	UserId       int             `gorm:"index;not null" json:"user_id"`                     // 用户id
	Symbol       string          `gorm:"size:255;index;not null" json:"symbol"`             // 交易对
	Amount       decimal.Decimal `gorm:"type:decimal(36,18);not null" json:"amount"`        // 数量
	FrozenAmount decimal.Decimal `gorm:"type:decimal(36,18);not null" json:"frozen_amount"` // 待成交数量
	CreateTime   time.Time       `gorm:"autoCreateTime;not null" json:"create_time"`        // 创建时间
	UpdateTime   time.Time       `gorm:"autoUpdateTime;not null" json:"update_time"`        // 更新时间
}

type WalletDao struct {
	db *gorm.DB
}

func NewWalletDao() *WalletDao {
	return &WalletDao{
		db: db,
	}
}
