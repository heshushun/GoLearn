package main

import "fmt"

type Stock struct {
	high float64
	low float64
	close float64
}

func modifyStock(stock *Stock) {
	stock.high = 475.10
	stock.low = 400.15
	stock.close = 450.75
}

func main() {
	goo := Stock{454.43, 421.01, 435.29}
	fmt.Println("Original Stock Data:", goo)
	modifyStock(&goo)
	fmt.Println("Modified Stock Data:", goo)
}