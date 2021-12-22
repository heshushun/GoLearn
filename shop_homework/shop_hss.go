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

type Shop struct {
	// {id: GridRaw}
	gridRows map[int]*GridRaw

	// {position: Grid}
	gridModels map[int]*Grid
}

func NewShop() *Shop {
	shop := &Shop{}
	return shop
}

func (r *Shop) Buy() {

}

func (r *Shop) Refresh() {

}

func (r *Shop) GetGridRaw(id int) *GridRaw {
	if r.gridRows == nil {
		r.gridRows = make(map[int]*GridRaw)
	}
	return r.gridRows[id]
}

type Grid struct {
	item        [2]int // (id, item)
	buyCount    int
	lastBuyTime int64
}

func (r *Grid) BuyItem() {

}

func (r *Grid) CanBuy() {

}

type GridRaw struct {
	id         int
	position   int
	item       int
	kind       int // 1: fixed 2: random
	randomProb int
	countLimit int // 0: 无限购买
	dayLimit   int // 单位: 天
}

func (r *GridRaw) GenItems(position int) [2]int {
	if r.kind == 1 {
		return [2]int{r.id, r.item}
	}

	return [2]int{}
}

func main() {

}
