package main

import (
	"fmt"
	"k8s-samples/cgo/qsort"
	"unsafe"
)

func main() {
	values := []int{42, 9, 101, 95, 27, 25}
	qsort.Sort(unsafe.Pointer(&values[0]),
		len(values), int(unsafe.Sizeof(values[0])),
		func(a, b unsafe.Pointer) int {
			pa, pb := (*int)(a), (*int)(b)
			return *pa - *pb
		},
	)
	fmt.Println(values)
}
