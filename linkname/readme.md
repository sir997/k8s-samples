## linkname
> 这个特殊的指令的作用域并不是紧跟的下一行代码，而是同一个包下生效。//go:linkname告诉 Go 的编译器把本地的(私有)变量或者方法localname链接到指定> 的变量或者方法importpath.name。简单来说，localname import.name指向的变量或者方法是同一个。因为这个指令破坏了类型系统和包的模块化原则，只有> 在引入 unsafe 包的前提下才能使用这个指令。

```
//go:linkname localname [importpath.name]
func hello(){
}
```

### Tips
如果出现：missing function body 在包内加一个 .s文件。因为go build默认加会加上-complete参数，加这个 .s 可绕开这个限制。