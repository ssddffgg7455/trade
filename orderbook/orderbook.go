package orderbook

import (
	"sync"
	"trade/dao"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/shopspring/decimal"
)

var orderBooks = make(map[string]*OrderBook)
var orderBooksMu sync.RWMutex

type OrderBook struct {
	bids *rbt.Tree // 买单列表（价格降序）
	asks *rbt.Tree // 卖单列表（价格升序）
	mu   sync.Mutex
}

func NewOrderBook(symbol string) *OrderBook {
	orderBooksMu.Lock()
	defer orderBooksMu.Unlock()

	if _, exists := orderBooks[symbol]; !exists {
		orderBooks[symbol] = &OrderBook{
			bids: rbt.NewWith(PriceComparator),
			asks: rbt.NewWith(PriceComparator),
		}
	}
	return orderBooks[symbol]
}

func GetOrderBook(symbol string) *OrderBook {
	orderBooksMu.RLock()
	ob, exists := orderBooks[symbol]
	orderBooksMu.RUnlock()

	if !exists {
		return NewOrderBook(symbol)
	}
	return ob
}

// AddOrder 添加订单到订单簿（挂单）
func (ob *OrderBook) AddOrder(order *dao.Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	switch order.Side {
	case dao.OrderSideBuy:
		ob.bids.Put(order.Id, order)
	case dao.OrderSideSell:
		ob.asks.Put(order.Id, order)
	}
}

// Match 尝试匹配订单（吃单）
func (ob *OrderBook) Match(order *dao.Order) []*dao.Trade {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	var trades []*dao.Trade
	var oppositeTree *rbt.Tree

	switch order.Side {
	case dao.OrderSideBuy:
		oppositeTree = ob.asks
	case dao.OrderSideSell:
		oppositeTree = ob.bids
	}

	remaining := order.Amount
	// 使用迭代器遍历红黑树
	it := oppositeTree.Iterator()
	if order.Side == dao.OrderSideBuy {
		it.Begin() // 卖单升序，从最小价格开始
	} else {
		it.End() // 买单降序，从最大价格开始
	}

	for remaining.IsPositive() {
		if order.Side == dao.OrderSideBuy && !it.Next() {
			break
		}
		if order.Side == dao.OrderSideSell && !it.Prev() {
			break
		}

		oppositeOrder := it.Value().(*dao.Order)
		if order.Type == dao.OrderTypeLimit {
			if (order.Side == dao.OrderSideBuy && order.Price.LessThan(oppositeOrder.Price)) ||
				(order.Side == dao.OrderSideSell && order.Price.GreaterThan(oppositeOrder.Price)) {
				break // 价格不匹配
			}
		}

		// 计算可成交数量
		available := oppositeOrder.Amount.Sub(oppositeOrder.Filled)
		fillAmount := decimal.Min(remaining, available)

		// 创建成交记录
		trade := &dao.Trade{
			TakerOrderId: order.Id,
			MakerOrderId: oppositeOrder.Id,
			Price:        oppositeOrder.Price,
			Amount:       fillAmount,
		}
		trades = append(trades, trade)

		// 更新订单状态
		remaining = remaining.Sub(fillAmount)
		oppositeOrder.Filled = oppositeOrder.Filled.Add(fillAmount)
		order.Filled = order.Filled.Add(fillAmount)

		// 检查是否完全成交
		if oppositeOrder.Filled.GreaterThanOrEqual(oppositeOrder.Amount) {
			// 移除完全成交的订单
			oppositeTree.Remove(oppositeOrder.Id)
		}
	}

	return trades
}

// RemoveOrder 取消订单
func (ob *OrderBook) RemoveOrder(orderId int, side dao.OrderSide) bool {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	switch side {
	case dao.OrderSideBuy:
		ob.bids.Remove(orderId)
	case dao.OrderSideSell:
		ob.asks.Remove(orderId)
	}

	return true
}

// Bids 获取买单列表（线程安全）
func (ob *OrderBook) Bids() []*dao.Order {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	keys := ob.bids.Keys()
	orders := make([]*dao.Order, 0, len(keys))
	for _, key := range keys {
		orders = append(orders, key.(*dao.Order))
	}
	return orders
}

// Asks 获取卖单列表（线程安全）
func (ob *OrderBook) Asks() []*dao.Order {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	keys := ob.asks.Keys()
	orders := make([]*dao.Order, 0, len(keys))
	for _, key := range keys {
		orders = append(orders, key.(*dao.Order))
	}
	return orders
}
