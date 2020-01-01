package main

import "fmt"

func main() {
	a := []int{1, 2, 3, 4, 5, 6}
	a = a[1:]  //删除开头1个元素
	a = a[2:]  //删除开头2个元素
	fmt.Println(a)

	//用append来删除 并且不移动数据指针
	b := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	b = append(b[:0], b[1:]...) // 删除开头1个元素
	b = append(b[:0], b[2:]...)  // 删除开头2个元素

	//中间位置删除
	b = append(b[:2], b[3:]...)

	//尾部删除
	b = b[:len(b)-1] //删除尾部1个元素
	b = b[:len(b)-2] //删除尾部2个元素

}
