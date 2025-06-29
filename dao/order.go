package dao

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type OrderType int

const (
	OrderTypeLimit  OrderType = 1 // 限价单
	OrderTypeMarket OrderType = 2 // 市价单
)

// OrderSide 订单方向
type OrderSide int

const (
	OrderSideBuy  OrderSide = 1 // 买入
	OrderSideSell OrderSide = 2 // 卖出
)

// OrderStatus 订单状态
type OrderStatus int

const (
	OrderStatusInit      OrderStatus = 1 // 新订单
	OrderStatusPartial   OrderStatus = 2 // 部分成交
	OrderStatusFilled    OrderStatus = 3 // 完全成交
	OrderStatusCancelled OrderStatus = 4 // 已取消
)

// Order 订单结构
type Order struct {
	Id         int             `gorm:"primaryKey;type:int;index;not null" json:"id"` // 订单id
	UserId     int             `gorm:"index;not null" json:"user_id"`                // 用户id
	Symbol     string          `gorm:"size:255;index;not null" json:"symbol"`        // 交易对
	Price      decimal.Decimal `gorm:"type:decimal(36,18);not null" json:"price"`    // 价格（市价单为0）
	Amount     decimal.Decimal `gorm:"type:decimal(36,18);not null" json:"amount"`   // 数量
	Filled     decimal.Decimal `gorm:"type:decimal(36,18);not null" json:"filled"`   // 已成交数量
	Type       OrderType       `gorm:"type:int;size:50;not null" json:"type"`        // 订单类型
	Side       OrderSide       `gorm:"type:int;size:50;not null" json:"side"`        // 订单方向
	Status     OrderStatus     `gorm:"type:int;size:50;not null" json:"status"`      // 订单状态
	CreateTime time.Time       `gorm:"autoCreateTime;not null" json:"create_time"`   // 创建时间
	UpdateTime time.Time       `gorm:"autoUpdateTime;not null" json:"update_time"`   // 更新时间
}

type OrderDao struct {
	db *gorm.DB
}

func NewOrderDao() *OrderDao {
	return &OrderDao{
		db: db,
	}
}

// CreateOrder 创建订单记录
func (o *OrderDao) CreateOrder(ctx context.Context, order *Order) error {
	return o.db.WithContext(ctx).Create(order).Error
}

// GetOrderById 根据订单ID获取订单记录
func (o *OrderDao) GetOrderById(ctx context.Context, orderId int) (*Order, error) {
	var order Order
	err := o.db.WithContext(ctx).Where("id = ?", orderId).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (o *OrderDao) GetOrderByUserId(ctx context.Context, userId int, symbol string, typ OrderType, side OrderSide, page, limit int) ([]*Order, error) {
	var orders []*Order
	offset := (page - 1) * limit
	db := o.db.WithContext(ctx).Where("user_id = ?", userId)
	if len(symbol) > 0 {
		db.Where("symbol = ?", symbol)
	}
	if typ > 0 {
		db.Where("type = ?", typ)
	}
	if side > 0 {
		db.Where("side = ?", side)
	}
	err := db.Offset(offset).Limit(limit).Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// UpdateOrderStatus 更新订单状态
func (o *OrderDao) UpdateOrderStatus(ctx context.Context, orderId int, status OrderStatus) error {
	return o.db.WithContext(ctx).Where("id = ?", orderId).Update("status", status).Error
}

// UpdateOrder 更新订单记录
func (o *OrderDao) UpdateOrder(ctx context.Context, order *Order) error {
	return o.db.WithContext(ctx).Updates(order).Error
}

// DeleteOrderByID 根据订单ID删除订单记录
func (o *OrderDao) DeleteOrderByID(ctx context.Context, orderId string) error {
	return o.db.WithContext(ctx).Where("id = ?", orderId).Delete(&Order{}).Error
}
