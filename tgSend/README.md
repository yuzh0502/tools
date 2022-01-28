# tgSend

## 使用方法

tgSend.Send("代理地址", chatID, "消息内容")

没有写不需要代理的逻辑，国内访问一般都是要代理的吧。

```go

package main

import (
	"fmt"
	"github.com/jlvihv/tools/tgSend"
)

func main() {
	err := tgSend.Send("http://localhost:7890", 956772010, "新的消息")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("发送成功")
}

```

### 获取chatID的方法

tg发送任意消息给@userinfobot，将返回chatID。
