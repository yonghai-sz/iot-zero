// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"flag"
	"fmt"
	"net/http"

	"iot-zero/services/platform-api/internal/authz"
	"iot-zero/services/platform-api/internal/config"
	"iot-zero/services/platform-api/internal/handler"
	"iot-zero/services/platform-api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/platform-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	var opts []rest.RunOption
	if c.PublicFiles != "" {
		opts = append(opts, rest.WithFileServer("/files", http.Dir(c.PublicFiles)))
	}
	server := rest.MustNewServer(c.RestConf, opts...)
	defer server.Stop()

	httpx.SetErrorHandler(authz.HTTPError)

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
