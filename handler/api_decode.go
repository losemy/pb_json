package handler

import (
	"encoding/json"
	"io"

	"pb_json/pb"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type Stream struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

func ApiDecode(r *ghttp.Request) {
	data, _ := io.ReadAll(r.Body)
	r.Response.Header().Set("Content-Type", "application/json")
	var stream *Stream
	if err := json.Unmarshal(data, &stream); err != nil {
		g.Log().Errorf(nil, "decode err")
		r.Response.Write(data)
		return
	}
	js, err := pb.Decode(stream.Data, nil)
	if err != nil {
		g.Log().Errorf(nil, "decode err: %v", err)
		r.Response.Write(data)
		return
	}
	g.Log().Infof(nil, "data -> result: %v -> %v", len(data), len(js))
	r.Response.Write(js)
}
