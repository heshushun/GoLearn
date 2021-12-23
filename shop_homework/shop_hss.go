package main

import (
	"fmt"
	"math/rand"
	"time"
)

const OneDay = 24 * time.Hour

func NowTime() int64 {
	return time.Now().Unix()
}

func AddDays(t int64, days int) int64 {
	return time.Unix(t, 0).AddDate(0, 0, days).Unix()
}

/*
*
* Role
*
**/
type Role struct {
	name  string
	level int
	vip   int
	shops map[string]*Shop // {shop_name: Shop}
}

func NewRole(name string, level, vip int) *Role {
	role := &Role{
		name:  name,
		level: level,
		vip:   vip,
	}
	role.shops = make(map[string]*Shop)
	return role
}

func (r *Role) addShop(name string, shop *Shop) {
	r.shops[name] = shop
}

/*
*
* Shop
*
**/
type Shop struct {
	// {id: GridRow}
	gridRows map[int]*GridRow

	gridRower *GridRower

	// {position: Grid}
	gridModels map[int]*Grid
}

func NewShop(rows []*GridRow) *Shop {
	shop := &Shop{}
	shop.gridRows = make(map[int]*GridRow)
	shop.gridRower = &GridRower{}
	shop.gridModels = make(map[int]*Grid)
	for _, row := range rows {
		shop.gridRows[row.id] = row
		shop.gridRower.Append(row.position, *row)
	}
	shop.gridRower.Done()
	return shop
}

func (r *Shop) Buy(role *Role, position, count int) {
	if grid, ok := r.gridModels[position]; ok {
		grid.BuyItem(role, count, func(id int) *GridRow {
			if r.gridRows == nil {
				r.gridRows = make(map[int]*GridRow)
			}
			return r.gridRows[id]
		})
	} else {
		fmt.Printf("no this position %d \n", position)
	}
}

func (r *Shop) Refresh(shopName string) {
	for _, position := range r.gridRower.positions {
		id, item := r.gridRower.Random(position)
		if id == 0 {
			continue
		}
		r.gridModels[position] = &Grid{
			id:          id,
			item:        item,
			buyCount:    0,
			lastBuyTime: 0,
		}
	}
	fmt.Printf("%s refresh success \n", shopName)
	r.ShowShop()
}

func (r *Shop) onCheckUnlock(role *Role, position int) {
	if grid, ok := r.gridModels[position]; ok {
		grid.Unlock(role, func(id int) *GridRow {
			if r.gridRows == nil {
				r.gridRows = make(map[int]*GridRow)
			}
			return r.gridRows[id]
		})
	} else {
		fmt.Printf("no this position %d \n", position)
	}
}

func (r *Shop) ShowShop() {
	fmt.Print("--------------------------\n")
	for pos, grid := range r.gridModels {
		fmt.Printf("%-2d ——> item: %-8s  buyCount: %-2d  lastTime: %-10v \n", pos, grid.item, grid.buyCount, grid.lastBuyTime)
	}
	fmt.Print("--------------------------\n\n")
}

/*
*
* Grid
*
**/
type Grid struct {
	id          int
	item        string
	buyCount    int
	lastBuyTime int64
}

func (r *Grid) BuyItem(role *Role, count int, f func(id int) *GridRow) {
	if !r.CanBuy(role, count, f) {
		return
	}

	fmt.Printf("%s buy %s %d success \n", role.name, r.item, count)

	r.buyCount += count
	r.lastBuyTime = NowTime()

}

func (r *Grid) CanBuy(role *Role, count int, f func(id int) *GridRow) bool {
	row := f(r.id)
	// level
	if role.level < row.LevelLimit {
		fmt.Printf("!!! %s buy %s %d fail level no enough \n", role.name, r.item, count)
		return false
	}

	// vip
	if role.vip < row.vipLimit {
		fmt.Printf("!!! %s buy %s %d fail vip no enough \n", role.name, r.item, count)
		return false
	}

	// count limit
	if row.countLimit != 0 {
		if r.buyCount+count > row.countLimit {
			fmt.Printf("!!! %s buy %s %d fail count no enough \n", role.name, r.item, count)
			return false
		}
	}

	// time limit
	if r.lastBuyTime != 0 {
		if row.dayLimit != 0 && AddDays(r.lastBuyTime, row.dayLimit) > NowTime() {
			fmt.Printf("!!! %s buy %s %d fail time limit \n", role.name, r.item, count)
			return false
		}
	}

	return true
}

func (r *Grid) Unlock(role *Role, f func(id int) *GridRow) {
	row := f(r.id)
	role.level = row.LevelLimit
	role.vip = row.vipLimit
}

/*
*
* GridRower
*
**/
type GridRower struct {
	rows      map[int][]GridRow // {position: row}
	positions []int
}

func (r *GridRower) Append(position int, row GridRow) {
	if r.rows == nil {
		r.rows = make(map[int][]GridRow)
	}
	if r.rows[position] == nil {
		r.rows[position] = make([]GridRow, 0, 8)
	}
	r.rows[position] = append(r.rows[position], row)
}

func (r *GridRower) Done() {
	positions := make([]int, 0, len(r.rows))
	for pos, _ := range r.rows {
		positions = append(positions, pos)
	}
	r.positions = positions
}

func (r *GridRower) Random(position int) (int, string) {
	if len(r.rows[position]) == 0 {
		return 0, ""
	}
	sum := 0
	// fixed
	for _, row := range r.rows[position] {
		if row.kind == 1 {
			return row.id, row.item
		}
		sum += row.randomWeight
	}

	// random
	rand.Seed(time.Now().Unix())
	rnd := rand.Intn(sum) + 1
	for _, row := range r.rows[position] {
		rnd -= row.randomWeight
		if rnd <= 0 {
			return row.id, row.item
		}
	}
	return 0, ""
}

/*
*
* GridRow
*
**/
type GridRow struct {
	id           int
	position     int
	item         string
	kind         int // 1: fixed 2: random
	randomWeight int
	countLimit   int // 0: 无限购买
	dayLimit     int // 0: 无时间限制 单位: 天
	LevelLimit   int // 等级限制
	vipLimit     int // vip限制
}

func main() {

	// role
	role := NewRole("hss", 10, 1)

	// gridRow
	row1 := GridRow{
		id:       1,
		position: 1,
		item:     "gold",
		kind:     1,
	}
	row2 := GridRow{
		id:         2,
		position:   2,
		item:       "rmb",
		kind:       1,
		countLimit: 1,
	}
	row3 := GridRow{
		id:       3,
		position: 3,
		item:     "box_1",
		kind:     1,
		dayLimit: 1,
	}
	row4 := GridRow{
		id:           4,
		position:     4,
		item:         "card_1",
		kind:         2,
		randomWeight: 10,
		countLimit:   1,
	}
	row5 := GridRow{
		id:           5,
		position:     4,
		item:         "card_2",
		kind:         2,
		randomWeight: 20,
		countLimit:   2,
	}
	row6 := GridRow{
		id:           4,
		position:     4,
		item:         "card_3",
		kind:         2,
		randomWeight: 30,
		countLimit:   3,
	}
	row7 := GridRow{
		id:         7,
		position:   5,
		item:       "exp",
		kind:       1,
		countLimit: 1,
		dayLimit:   1,
		LevelLimit: 20,
	}

	// gen shop
	rowLen := 7
	rows := make([]*GridRow, 0, rowLen)
	rows = append(rows, &row1)
	rows = append(rows, &row2)
	rows = append(rows, &row3)
	rows = append(rows, &row4)
	rows = append(rows, &row5)
	rows = append(rows, &row6)
	rows = append(rows, &row7)
	shop1 := NewShop(rows)
	role.addShop("shop_1", shop1)

	// refresh shop
	shop := role.shops["shop_1"]
	shop.Refresh("shop_1")

	// buy shop
	shop.Buy(role, 1, 1)
	shop.ShowShop()
	shop.Buy(role, 5, 1)
	shop.ShowShop()
	shop.onCheckUnlock(role, 5)
	shop.Buy(role, 5, 1)
	shop.ShowShop()
	shop.Buy(role, 5, 1)
	shop.ShowShop()

}
