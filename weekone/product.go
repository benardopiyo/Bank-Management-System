package main

import "fmt"

type Product struct {
	name       string
	color      string
	price      int
	tax        int
	totalPrice int
}

func TotalPrice() {
	var price int
	tax := tax / 100 * price

	totalPrice := price + tax

	fmt.Println(totalPrice)
}
