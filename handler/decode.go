package handler

import (
	"io"

	"pb_json/pb"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func Decode(r *ghttp.Request) {
	data, _ := io.ReadAll(r.Body)
	js, err := pb.Decode(data, nil)
	if err != nil {
		g.Log().Errorf(nil, "decode err: %v", err)
		r.Response.Write(data)
		return
	}
	g.Log().Infof(nil, "data -> result: %v -> %v", len(data), len(js))
	r.Response.Write(js)
}
