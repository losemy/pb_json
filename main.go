package main

import (
	"context"
	"fmt"

	"pb_json/handler"

	"github.com/gogf/gf/v2/frame/g"
)

func main() {
	s := g.Server()

	s.BindHandler("/decode", handler.Decode)
	s.BindHandler("/api_decode", handler.ApiDecode)

	port := g.Cfg().MustGet(context.Background(), "port")
	s.SetPort(port.Int())
	// 数据转换成对应的结构
	fmt.Println(g.Cfg().MustData(context.Background()))
	s.Run()
}
