package main

type Role struct {
	level int
	vip   int
	shops map[string]*Shop // {shop_name: Shop}
}

func NewRole(level, vip int) *Role {
	role := &Role{
		level: level,
		vip:   vip,
	}
	return role
}

func (r *Role) addShop(name string, shop *Shop) {
	r.shops[name] = shop
}

type ShopGridInfo struct {
	items       [2]int // (id, item)
	buyCount    int
	lastBuyTime int64
}

type Shop struct {
	// {id: grid}
	grids map[int]*Grid

	// {position: gridModel}
	gridModels map[int]*ShopGridInfo
}

func NewShop() *Shop {
	shop := &Shop{}
	return shop
}

func (r *Shop) Buy() {

}

func (r *Shop) Refresh() {

}

type Grid struct {
	position   int
	item       string
	kind       int // 1: fixed 2: random
	countLimit int
	dayLimit   int // 单位: 天
}

func (r *Grid) BuyItem() {

}

func (r *Grid) CanBuy() {

}

func main() {

}
