package dao

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Trade 成交记录
type Trade struct {
	Id           int             `gorm:"primaryKey;autoIncrement;type:int;index;not null" json:"id"` // 交易记录id
	TakerOrderId int             `gorm:"type:int;index;not null" json:"taker_order_id"`              // 吃单订单id
	MakerOrderId int             `gorm:"type:int;;index;not null" json:"maker_order_id"`             // 挂单订单id
	Price        decimal.Decimal `gorm:"type:decimal(36,18);not null" json:"price"`                  // 成交价格
	Amount       decimal.Decimal `gorm:"type:decimal(36,18);not null" json:"amount"`                 // 成交数量
	Timestamp    time.Time       `gorm:"autoCreateTime;not null" json:"timestamp"`                   // 成交时间
}

type TradeDao struct {
	db *gorm.DB
}

func NewTradeDao() *TradeDao {
	return &TradeDao{
		db: db,
	}
}

// CreateTrade 创建成交记录
func (t *TradeDao) CreateTrade(ctx context.Context, trade *Trade) error {
	return t.db.WithContext(ctx).Create(trade).Error
}

// CreateTrades 批量创建成交记录
func (t *TradeDao) CreateTrades(ctx context.Context, trades []*Trade) error {
	return t.db.WithContext(ctx).Create(trades).Error
}

// GetTradesByTakerOrderId 根据吃单订单ID获取成交记录，支持分页
func (t *TradeDao) GetTradesByTakerOrderId(ctx context.Context, takerOrderId string, page, limit int) ([]Trade, error) {
	var trades []Trade
	// 计算偏移量
	offset := (page - 1) * limit
	err := t.db.WithContext(ctx).Where("taker_order_id = ?", takerOrderId).Offset(offset).Limit(limit).Find(&trades).Error
	if err != nil {
		return nil, err
	}
	return trades, nil
}

// GetTradesByMakerOrderId 根据挂单订单ID获取成交记录，支持分页
func (t *TradeDao) GetTradesByMakerOrderId(ctx context.Context, makerOrderId string, page, limit int) ([]Trade, error) {
	var trades []Trade
	// 计算偏移量
	offset := (page - 1) * limit
	err := t.db.WithContext(ctx).Where("maker_order_id = ?", makerOrderId).Offset(offset).Limit(limit).Find(&trades).Error
	if err != nil {
		return nil, err
	}
	return trades, nil
}

// GetAllTrades 获取所有成交记录，支持分页
func (t *TradeDao) GetAllTrades(ctx context.Context, page, limit int) ([]Trade, error) {
	var trades []Trade
	// 计算偏移量
	offset := (page - 1) * limit
	err := t.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&trades).Error
	if err != nil {
		return nil, err
	}
	return trades, nil
}
