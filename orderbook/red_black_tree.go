package orderbook

import (
	"trade/dao"

	rbt "github.com/emirpasic/gods/trees/redblacktree"
)

type RedBlackTree *rbt.Tree

func NewRedBlackTree() *rbt.Tree {
	return rbt.NewWith(PriceComparator)
}

// PriceComparator 价格比较器
func PriceComparator(a, b interface{}) int {
	aOrder := a.(*dao.Order)
	bOrder := b.(*dao.Order)
	switch {
	case aOrder.Price.GreaterThan(bOrder.Price):
		return 1
	case aOrder.Price.LessThan(bOrder.Price):
		return -1
	default:
		return 0
	}
}
