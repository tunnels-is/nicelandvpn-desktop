package main

import "fmt"

func main() {
	// core.GetOpenSockets()
	X := byte(1)
	x1 := X << 7
	x2 := x1 >> 7
	fmt.Printf("%08b\n", X)
	fmt.Printf("%08b\n", x1)
	fmt.Printf("%08b\n", x2)

	Y := byte(252)
	y1 := Y & 0x07 >> 2
	// y2 := y1 >> 2

	fmt.Printf("%08b\n", Y)
	fmt.Printf("%08b\n", y1)
	// fmt.Printf("%08b\n", y2)
	// fmt.Printf("%08b\n", X>>7)
}
