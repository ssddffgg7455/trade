package matchingengine

import (
	"context"
	"sync"

	"trade/dao"
	"trade/orderbook"
)

// 可hash拆分 或按币种拆分
var matchingEngines = make(map[string]*MatchingEngine)
var matchingEnginesMu sync.RWMutex

type MatchingEngine struct {
	orderBook       *orderbook.OrderBook
	orderChan       chan *dao.Order
	cancelOrderChan chan int
	stopChan        chan struct{}
	wg              sync.WaitGroup
	symbol          string // 标记该引擎对应的交易对
}

func NewMatchingEngine(symbol string) *MatchingEngine {
	matchingEnginesMu.Lock()
	defer matchingEnginesMu.Unlock()

	if _, exists := matchingEngines[symbol]; !exists {
		matchingEngines[symbol] = &MatchingEngine{
			orderBook:       orderbook.NewOrderBook(symbol),
			orderChan:       make(chan *dao.Order, 10000), // 需要监控订单队列
			cancelOrderChan: make(chan int, 10000),
			stopChan:        make(chan struct{}),
			symbol:          symbol,
		}
		matchingEngines[symbol].Start()
	}
	return matchingEngines[symbol]
}

func GetMatchingEngine(symbol string) *MatchingEngine {
	matchingEnginesMu.RLock()
	engine, exists := matchingEngines[symbol]
	matchingEnginesMu.RUnlock()

	if !exists {
		return NewMatchingEngine(symbol)
	}
	return engine
}

func MatchingEngineClose() {
	matchingEnginesMu.Lock()
	defer matchingEnginesMu.Unlock()
	for _, engine := range matchingEngines {
		engine.Stop()
	}
}

// Start 启动匹配引擎
func (e *MatchingEngine) Start() {
	e.wg.Add(1)
	go e.processOrders()
}

// Stop 停止匹配引擎
func (e *MatchingEngine) Stop() {
	close(e.stopChan)
	e.wg.Wait()
}

// SubmitOrder 提交订单到匹配引擎
func (e *MatchingEngine) SubmitOrder(order *dao.Order) {
	if order.Symbol == e.symbol {
		e.orderChan <- order
	}
}

// SubmitCancelOrder 提交取消订单到匹配引擎
func (e *MatchingEngine) SubmitCancelOrder(orderId int) {
	e.cancelOrderChan <- orderId
}

// OrderBook 获取订单簿
func (e *MatchingEngine) OrderBook() *orderbook.OrderBook {
	return e.orderBook
}

// processOrders 处理订单的核心循环
func (e *MatchingEngine) processOrders() {
	defer e.wg.Done()

	for {
		select {
		case <-e.stopChan:
			return
		case order := <-e.orderChan:
			e.processOrder(order)
		case orderId := <-e.cancelOrderChan:
			e.removeOrder(orderId)
		}
	}
}

// processOrder 处理单个订单
func (e *MatchingEngine) processOrder(order *dao.Order) {
	switch order.Type {
	case dao.OrderTypeLimit:
		e.handleLimitOrder(order)
	case dao.OrderTypeMarket:
		e.handleMarketOrder(order)
	}
}

// handleLimitOrder 处理限价单（挂单）
func (e *MatchingEngine) handleLimitOrder(order *dao.Order) {
	// 先尝试匹配（可能部分成交）
	trades := e.orderBook.Match(order)

	// 如果有剩余数量，添加到订单簿
	if order.Filled.LessThan(order.Amount) {
		order.Status = dao.OrderStatusPartial
		e.orderBook.AddOrder(order)
	} else {
		order.Status = dao.OrderStatusFilled
	}

	// 更新订单信息
	err := dao.NewOrderDao().UpdateOrder(context.Background(), order)
	if err != nil {
		// TODO: 处理错误
		return
	}

	// TODO: 更新钱包

	// 批量创建成交记录
	err = dao.NewTradeDao().CreateTrades(context.Background(), trades)
	if err != nil {
		// TODO: 处理错误
		return
	}

	// TODO: 发送交易结果
	_ = trades
}

// handleMarketOrder 处理市价单（吃单）
func (e *MatchingEngine) handleMarketOrder(order *dao.Order) {
	trades := e.orderBook.Match(order)

	// 如果有剩余数量，撤单
	if order.Filled.LessThan(order.Amount) {
		order.Status = dao.OrderStatusPartial
	} else {
		order.Status = dao.OrderStatusFilled
	}

	// 更新订单信息
	err := dao.NewOrderDao().UpdateOrder(context.Background(), order)
	if err != nil {
		// TODO: 处理错误
		return
	}

	// TODO: 更新钱包

	// 批量创建成交记录
	err = dao.NewTradeDao().CreateTrades(context.Background(), trades)
	if err != nil {
		// TODO: 处理错误
		return
	}

	// TODO: 发送交易结果
	_ = trades
}

func (e *MatchingEngine) removeOrder(orderId int) {
	order, err := dao.NewOrderDao().GetOrderById(context.Background(), orderId)
	if err != nil {
		// TODO: 处理错误
		return
	}

	succ := e.orderBook.RemoveOrder(orderId, order.Side)

	// TODO: 发送撤单结果
	_ = succ
}
