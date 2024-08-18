package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/valyala/fasthttp"
)

var (
	proxy_headers = []string{
		"Forwarded", "Proxy-Authorization",
		"X-Forwarded-For", "Proxy-Authenticate",
		"X-Requested-With", "From",
		"X-Real-Ip", "Via", "True-Client-Ip", "Proxy_Connection",
	}

	port  = os.Getenv("PORT")
	token = []byte(os.Getenv("TOKEN"))
)

func fastHTTPHandler(ctx *fasthttp.RequestCtx) {

	if !bytes.Equal(ctx.Path(), token) {
		ctx.Response.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	real := ctx.Request.URI().QueryArgs().Peek("real")
	remoteIP := ctx.RemoteIP().String()

	if len(real) > 0 {
		if string(real) == remoteIP {
			json(ctx, 4, "Direct")
			return
		}
		remoteIP = string(real)
	}

	for _, header := range proxy_headers {
		if v := ctx.Request.Header.Peek(header); v != nil {
			if string(v) == remoteIP {
				json(ctx, 3, "Transparent")
				return
			}

			json(ctx, 2, "Anonymous")
			return
		}
	}

	json(ctx, 1, "Elite")
}

func json(ctx *fasthttp.RequestCtx, level int, name string) {
	log.Printf("Request from %s is %s", ctx.RemoteIP().String(), name)
	ctx.Response.Header.Set("Content-Type", "application/json")
	fmt.Fprintf(ctx, `{"level":%d,"name":"%s"}`, level, name)
}

func main() {
	if port == "" {
		port = "8080"
	}

	token = []byte("/" + string(token))

	fmt.Printf("Listening on :%s%s", port, token)

	h := fasthttp.CompressHandler(fastHTTPHandler)

	if err := fasthttp.ListenAndServe(":"+port, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %v", err)
	}
}
