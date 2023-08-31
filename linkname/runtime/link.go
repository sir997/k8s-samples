package runtime

import (
	"fmt"

	_ "unsafe"
)

//go:linkname runtime_link k8s-samples/linkname/link.Runtime_link
func runtime_link() {
	fmt.Println("runtime link")
}
