package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

/*
*
* 作业 设计解释
* 本次作业 实际上实现三大功能: 刷新(Refresh)、购买(Buy)、解锁(Unlock)。
* 2个类层面: 商店(Shop)、栏位格子(Grid)。
* 2大数据源: 配表(GridRows)、格子model数据(GridModel)。
*
* 我的刷新业务逻辑是 一栏位多id的随机构建。所以设计上这块我并没有放到格子层面，也没有做过多的设计。只是包装了一个GridRower 用来管理 GridRow。
* 设计层面:
* 1、商店(Shop) 设计很轻量。只是一个盒子，主要为 格子的组合、提供数据源。只对外提供三个功能 刷新(Refresh)、购买(Buy)、解锁(Unlock)的入口。
* 2、设计主要体现在 栏位格子(Grid)上。抽象出格子基类Grid 以及通用接口IGrid 只关心 购买(Buy)和解锁(Unlock), 实体类逻辑是对各自购买限制种类格子的通用接口实现。
* 3、单一职责原则: 如果某类购买限制种类格子的业务发生改变，只需要改动自身的业务实现，不会影响到其他种类。
* 4、开闭原则: 如果需要新增一新类格子，只需要添加一个实现类实现通用接口，不用动原有框架较易扩展。
*
**/

const (
	LevelGrid = 1
	LimitGrid = 2
	MaryGrid  = 3
)

/*
*
* Role
* 注释：纯属为了游戏业务上更舒服理解而顺便添加的对象类。
* 		刚好还可以用它来 设置其中一些购买限制条件数据源。
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

func (r *Role) printShopNames() {
	fmt.Print("--------------------------\n")
	for name, _ := range r.shops {
		fmt.Printf("——> %s\n", name)
	}
	fmt.Print("--------------------------\n\n")
}

/*
*
* Shop
* 注释：a. 两大数据源 和 配置管理者gridRower 构成。
* 		b. 对外提供刷新(Refresh)、购买(Buy)、解锁(OnCheckUnlock) 入口。
* 		c. 职责仅是 组织格子。
*		d. 它不关心 怎么刷新 怎么购买 怎么解锁，都是格子自己的事。不关心是什么种类格子用怎么购买逻辑，只要你是Grid就好。
*
**/
type Shop struct {
	gridRower *GridRower // GridRower 负责构建管理gridRow

	// {id: GridRow}
	gridRows map[int]*GridRow // gridRow数据

	// {position: Grid}
	gridModels map[int]*Grid // GridModel数据
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
		grid.BuyItem(role, count)
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
		row := r.gridRows[id]
		gridModel := &GridModel{
			id:          id,
			item:        item,
			buyCount:    0,
			lastBuyTime: 0,
		}
		grid := NewGrid(GridFactory(row.classType), gridModel, row)
		r.gridModels[position] = grid
	}
	fmt.Printf("%s refresh success \n", shopName)
	r.ShowShop()
}

func (r *Shop) OnCheckUnlock(role *Role, position int) {
	if grid, ok := r.gridModels[position]; ok {
		grid.Unlock(role)
	} else {
		fmt.Printf("no this position %d \n", position)
	}
}

// 打印显示商店数据。（交互作用 与业务逻辑无关）
func (r *Shop) ShowShop() {
	fmt.Print("--------------------------\n")
	newPos := make([]int, len(r.gridRower.positions))
	copy(newPos, r.gridRower.positions)
	sort.Ints(newPos)
	for _, pos := range newPos {
		if grid, ok := r.gridModels[pos]; ok {
			row := r.gridRows[grid.Row.id]
			fmt.Printf("%-2d ——> item: %-8s buyLimit: %-2d buyCount: %-2d classType: %-2d dayLimit: %-2d lastTime: %-10v roleLevel: %-2d vipLevel: %-2d \n",
				pos, grid.DB.item, row.countLimit, grid.DB.buyCount, row.classType, row.dayLimit, grid.DB.lastBuyTime, row.roleLevel, row.vipLevel)
		}
	}
	fmt.Print("--------------------------\n\n")
}

/*
*
* Grid
* 注释：a. Grid 由 IGrid逻辑接口 和 GridModel数据 构成。
* 		b. Grid 对外也只提供通用函数 购买（BuyItem）、解锁（Unlock）。因为对外shop也只关心这两个。
* 		c. GridModel数据源由 shop 提供。
*		d. Grid是base类，抽象出共用IGrid接口，每个具体Grid类只要实现IGrid接口。
*		e. GridLevelVip、GridLimit、GridMany 三个是每个的具体格子类。（当然这么拆分不严谨，随便拆的）
*
**/
func GridFactory(gridType int) IGrid {
	switch gridType {
	case LevelGrid:
		return NewGridLevelVip()
	case LimitGrid:
		return NewGridLimit()
	case MaryGrid:
		return NewGridMany()
	default:
		fmt.Printf("%d classType error \n", gridType)
		return nil
	}
}

type IGrid interface {
	CanBuy(role *Role, count int, row *GridRow, m *GridModel) bool
	DoUnlock(role *Role, row *GridRow, m *GridModel)
}

// GridModel数据结构
type GridModel struct {
	id          int
	item        string
	buyCount    int
	lastBuyTime int64
}

// Grid基类
type Grid struct {
	IGrid
	DB  *GridModel
	Row *GridRow
}

func NewGrid(iGrid IGrid, db *GridModel, row *GridRow) *Grid {
	grid := Grid{
		IGrid: iGrid,
		DB:    db,
		Row:   row,
	}
	return &grid
}

func (r *Grid) BuyItem(role *Role, count int) {
	if !r.CanBuy(role, count, r.Row, r.DB) {
		return
	}

	fmt.Printf("%s buy %s %d success \n", role.name, r.DB.item, count)

	r.DB.buyCount += count
	r.DB.lastBuyTime = NowTime()

}

func (r *Grid) Unlock(role *Role) {
	if r.Row.vipLevel != 0 || r.Row.roleLevel != 0 {
		r.DoUnlock(role, r.Row, r.DB)
	}
	if r.Row.dayLimit != 0 || r.Row.countLimit != 0 {
		r.DoUnlock(role, r.Row, r.DB)
	}
}

/*
*
* GridLevelVip  具体格子类一
*
**/
type GridLevelVip struct {
}

func NewGridLevelVip() *GridLevelVip {
	return &GridLevelVip{}
}

func (r *GridLevelVip) DoUnlock(role *Role, row *GridRow, m *GridModel) {
	role.level = row.roleLevel
	role.vip = row.vipLevel
}

func (r *GridLevelVip) CanBuy(role *Role, count int, row *GridRow, m *GridModel) bool {
	// level
	if role.level < row.roleLevel {
		fmt.Printf("!!! %s buy %s %d fail level no enough \n", role.name, m.item, count)
		return false
	}

	// vip
	if role.vip < row.vipLevel {
		fmt.Printf("!!! %s buy %s %d fail vip no enough \n", role.name, m.item, count)
		return false
	}
	return true
}

/*
*
* GridLimit  具体格子类二
*
**/
type GridLimit struct {
}

func NewGridLimit() *GridLimit {
	return &GridLimit{}
}

func (r *GridLimit) DoUnlock(role *Role, row *GridRow, m *GridModel) {
	m.buyCount = 0
	m.lastBuyTime = 0
}

func (r *GridLimit) CanBuy(role *Role, count int, row *GridRow, m *GridModel) bool {
	// count limit
	if row.countLimit != 0 {
		if m.buyCount+count > row.countLimit {
			fmt.Printf("!!! %s buy %s %d fail count no enough \n", role.name, m.item, count)
			return false
		}
	}

	// time limit
	if m.lastBuyTime != 0 {
		if row.dayLimit != 0 && AddDays(m.lastBuyTime, row.dayLimit) > NowTime() {
			fmt.Printf("!!! %s buy %s %d fail time limit \n", role.name, m.item, count)
			return false
		}
	}

	return true
}

/*
*
* GridMany  具体格子类三
*
**/
type GridMany struct {
}

func NewGridMany() *GridMany {
	return &GridMany{}
}

func (r *GridMany) DoUnlock(role *Role, row *GridRow, m *GridModel) {
	role.level = row.roleLevel
	role.vip = row.vipLevel
	m.buyCount = 0
	m.lastBuyTime = 0
}

func (r *GridMany) CanBuy(role *Role, count int, row *GridRow, m *GridModel) bool {
	// level
	if role.level < row.roleLevel {
		fmt.Printf("!!! %s buy %s %d fail level no enough \n", role.name, m.item, count)
		return false
	}

	// vip
	if role.vip < row.vipLevel {
		fmt.Printf("!!! %s buy %s %d fail vip no enough \n", role.name, m.item, count)
		return false
	}

	// count limit
	if row.countLimit != 0 {
		if m.buyCount+count > row.countLimit {
			fmt.Printf("!!! %s buy %s %d fail count no enough \n", role.name, m.item, count)
			return false
		}
	}

	// time limit
	if m.lastBuyTime != 0 {
		if row.dayLimit != 0 && AddDays(m.lastBuyTime, row.dayLimit) > NowTime() {
			fmt.Printf("!!! %s buy %s %d fail time limit \n", role.name, m.item, count)
			return false
		}
	}

	return true
}

/*
*
* GridRower
* 注释：a. 配置GridRow的管理者。
*		b. 对GridRow做个二次处理化并缓存。类似于ClassInit。
* 		c. 一个栏位多个row，因为刷新构建就它处理了。
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
* 注释：配置结构
*
**/
type GridRow struct {
	id           int
	position     int
	item         string
	kind         int // 控制刷新 1: fixed 2: random
	randomWeight int
	classType    int // 控制购买 1: 等级 2: limit 3: mary （注: 格子购买限制分类 3是1和2之和）
	countLimit   int // 0: 无限购买
	dayLimit     int // 0: 无时间限制 单位: 天
	roleLevel    int // 等级限制
	vipLevel     int // vip限制
}

func main() {

	// role
	role := NewRole("hss", 10, 1)

	// gridRow（TODO kind 控制刷新；classType 控制购买）
	row1 := GridRow{id: 1, position: 1, item: "gold", kind: 1, classType: 1}
	row2 := GridRow{id: 2, position: 2, item: "rmb", kind: 1, classType: 2, countLimit: 1}
	row3 := GridRow{id: 3, position: 3, item: "box_1", kind: 1, classType: 2, dayLimit: 1}
	row4 := GridRow{id: 4, position: 4, item: "card_1", kind: 2, classType: 2, randomWeight: 10, countLimit: 1}
	row5 := GridRow{id: 5, position: 4, item: "card_2", kind: 2, classType: 2, randomWeight: 20, countLimit: 2}
	row6 := GridRow{id: 4, position: 4, item: "card_3", kind: 2, classType: 2, randomWeight: 30, countLimit: 3}
	row7 := GridRow{id: 7, position: 5, item: "exp", kind: 1, classType: 3, countLimit: 2, dayLimit: 1, roleLevel: 20}
	row8 := GridRow{id: 8, position: 6, item: "coin3", kind: 1, classType: 3, countLimit: 1, dayLimit: 1, vipLevel: 7}

	// gen shop
	rowLen := 8
	rows := make([]*GridRow, 0, rowLen)
	rows = append(rows, &row1)
	rows = append(rows, &row2)
	rows = append(rows, &row3)
	rows = append(rows, &row4)
	rows = append(rows, &row5)
	rows = append(rows, &row6)
	rows = append(rows, &row7)
	rows = append(rows, &row8)

	shop1 := NewShop(rows)
	shop1.Refresh("shop_1")
	role.addShop("shop_1", shop1)

	shop2 := NewShop(rows)
	shop2.Refresh("shop_2")
	role.addShop("shop_2", shop2)

	shop := role.shops["shop_1"]
	// buy shop
	shop.Buy(role, 1, 1)
	shop.ShowShop()
	shop.Buy(role, 1, 1)
	shop.ShowShop()
	shop.Buy(role, 2, 2)
	shop.ShowShop()
	shop.Buy(role, 2, 1)
	shop.ShowShop()
	shop.Buy(role, 5, 1)
	shop.ShowShop()
	shop.OnCheckUnlock(role, 5)
	shop.Buy(role, 5, 1)
	shop.ShowShop()
	shop.Buy(role, 5, 1)
	shop.ShowShop()

	// cmd 命令操作
	CmdInput(role)

}

// 命令行交互函数（交互作用 与业务逻辑无关）
func CmdInput(role *Role) {
	for {
		PrintPrompt()
		buffer := ReadInput()

		if buffer == "" {
			continue
		}

		if buffer == "exit" {
			os.Exit(0)
		} else if buffer == "help" {
			PrintHelp()
		} else if buffer == "shop" {
			role.printShopNames()
		} else {
			cmdList := strings.Fields(buffer)
			if len(cmdList) < 2 {
				fmt.Printf("Cmd error. Could not parse.\n")
				continue
			}
			cmd := strings.TrimSpace(cmdList[0])
			shopName := strings.TrimSpace(cmdList[1])
			shop, ok := role.shops[shopName]
			if !ok {
				fmt.Printf("Cmd shop name error. Could not parse.\n")
			}
			switch cmd {
			case "show":
				shop.ShowShop()
			case "refresh":
				shop.Refresh(shopName)
			case "unlock":
				if len(cmdList) < 3 {
					fmt.Printf("Cmd error. Could not parse.\n")
					continue
				}
				position, err := strconv.Atoi(strings.TrimSpace(cmdList[2]))
				if err != nil {
					fmt.Printf("Cmd pos error. Could not parse.\n")
				}
				shop.OnCheckUnlock(role, position)
			case "buy":
				if len(cmdList) < 4 {
					fmt.Printf("Cmd error. Could not parse.\n")
					continue
				}
				position, err1 := strconv.Atoi(strings.TrimSpace(cmdList[2]))
				if err1 != nil {
					fmt.Printf("Cmd pos error. Could not parse.\n")
				}
				count, err2 := strconv.Atoi(strings.TrimSpace(cmdList[3]))
				if err2 != nil {
					fmt.Printf("Cmd count error. Could not parse.\n")
				}
				shop.Buy(role, position, count)
			default:
				fmt.Printf("Cmd error. Could not parse.\n")
				continue
			}
		}
	}
}

// 打印输入提示（交互作用 与业务逻辑无关）
func PrintPrompt() {
	fmt.Printf("cmd(注: help) > ")
}

// 打印Help（交互作用 与业务逻辑无关）
func PrintHelp() {
	fmt.Print("--------------------------\n")
	fmt.Printf("%-20s :退出 \n", "exit")
	fmt.Printf("%-20s :商店列表 \n", "shop")
	fmt.Printf("%-20s :商店信息 \n", "show shop_1 ")
	fmt.Printf("%-20s :商店解锁 \n", "unlock shop_1 6")
	fmt.Printf("%-20s :商店刷新 \n", "refresh shop_1")
	fmt.Printf("%-20s :商店购买 \n", "buy shop_1 2 2")
	fmt.Print("--------------------------\n\n")
}

// 接收输入读取（交互作用 与业务逻辑无关）
func ReadInput() string {
	reader := bufio.NewReader(os.Stdin)
	res, _, err := reader.ReadLine()
	buffer := strings.TrimSpace(string(res))
	if err != nil {
		fmt.Printf("Error reading input %v \n", err)
		os.Exit(0)
	}
	return buffer
}

const OneDay = 24 * time.Hour

func NowTime() int64 {
	return time.Now().Unix()
}

func AddDays(t int64, days int) int64 {
	return time.Unix(t, 0).AddDate(0, 0, days).Unix()
}
