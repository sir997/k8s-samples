package main

// mac可以使用go:build unix，报红是goland不识别
// file_unix.go file_darwin.go中不能存在同名函数/变量等，会报重复
func main() {
	help()
}
