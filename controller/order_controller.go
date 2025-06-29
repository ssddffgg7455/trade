package controller

import (
	"errors"
	"net/http"
	"time"

	"trade/dao"
	"trade/matchingengine"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
)

type OrderController struct {
	BaseController
}

func (o *OrderController) Router(router *gin.Engine) {
	entry := router.Group("/order")
	entry.POST("/submit", o.SubmitOrder)
	entry.PUT("/cancel/:orderId", o.CancelOrder)
	entry.GET("/:orderId", o.GetOrder)
	entry.GET("", o.GetOrderList)
}

type SubmitOrderReq struct {
	UserId int             `json:"user_id" form:"user_id" binding:"required"`
	Symbol string          `json:"symbol" form:"symbol" binding:"required"`
	Price  decimal.Decimal `json:"price" form:"price"`
	Amount decimal.Decimal `json:"amount" form:"amount"`
	Type   dao.OrderType   `json:"type" form:"type"`
	Side   dao.OrderSide   `json:"side" form:"side"`
}

func (o *OrderController) SubmitOrder(c *gin.Context) {
	ctx := o.BaseContext(c)
	req := new(SubmitOrderReq)
	err := c.ShouldBind(req)
	if err != nil {
		respond(c, http.StatusBadRequest, nil, err)
		return
	}

	order := &dao.Order{
		UserId:     req.UserId,
		Symbol:     req.Symbol,
		Price:      req.Price,
		Amount:     req.Amount,
		Type:       req.Type,
		Side:       req.Side,
		CreateTime: time.Now(),
	}

	err = dao.NewOrderDao().CreateOrder(ctx, order)
	if err != nil {
		respond(c, http.StatusInternalServerError, req, err)
		return
	}

	go matchingengine.GetMatchingEngine(order.Symbol).SubmitOrder(order)

	respond(c, http.StatusOK, nil, nil)
}

func (o *OrderController) CancelOrder(c *gin.Context) {
	ctx := o.BaseContext(c)
	orderId := cast.ToInt(c.Param("orderId"))

	order, err := dao.NewOrderDao().GetOrderById(ctx, orderId)
	if err != nil {
		respond(c, http.StatusInternalServerError, orderId, err)
		return
	}
	if order.Id == 0 {
		err := errors.New("order not exist")
		respond(c, http.StatusBadRequest, orderId, err)
		return
	}

	if order.Status != dao.OrderStatusInit {
		err := errors.New("order status not init")
		respond(c, http.StatusBadRequest, orderId, err)
		return
	}

	err = dao.NewOrderDao().UpdateOrderStatus(ctx, orderId, dao.OrderStatusCancelled)
	if err != nil {
		respond(c, http.StatusInternalServerError, orderId, err)
		return
	}

	go matchingengine.GetMatchingEngine(order.Symbol).SubmitCancelOrder(orderId)

	respond(c, http.StatusOK, orderId, nil)
}

func (o *OrderController) GetOrder(c *gin.Context) {
	ctx := o.BaseContext(c)
	orderId := cast.ToInt(c.Param("orderId"))

	order, err := dao.NewOrderDao().GetOrderById(ctx, orderId)
	if err != nil {
		respond(c, http.StatusInternalServerError, orderId, err)
		return
	}
	respond(c, http.StatusOK, order, nil)
}

type GetOrderListReq struct {
	UserId int           `json:"user_id" form:"user_id" binding:"required"`
	Symbol string        `json:"symbol" form:"symbol"`
	Type   dao.OrderType `json:"type" form:"type"`
	Side   dao.OrderSide `json:"side" form:"side"`
	Page   int           `json:"page" form:"page"`
	Limit  int           `json:"limit" form:"limit"`
}

func (o *OrderController) GetOrderList(c *gin.Context) {
	ctx := o.BaseContext(c)
	req := new(GetOrderListReq)
	err := c.ShouldBind(req)
	if err != nil {
		respond(c, http.StatusBadRequest, nil, err)
		return
	}

	if req.UserId == 0 {
		err := errors.New("user_id is required")
		respond(c, http.StatusBadRequest, nil, err)
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 100
	}

	orders, err := dao.NewOrderDao().GetOrderByUserId(ctx, req.UserId, req.Symbol, req.Type, req.Side, req.Page, req.Limit)
	if err != nil {
		respond(c, http.StatusInternalServerError, req, err)
		return
	}
	respond(c, http.StatusOK, orders, nil)
}
