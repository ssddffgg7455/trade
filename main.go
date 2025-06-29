package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"trade/controller"
	"trade/dao"
	"trade/matchingengine"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func main() {
	app := gin.Default()

	pprof.Register(app)

	dao.Init()

	new(controller.OrderController).Router(app)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      app,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}

	go server.ListenAndServe()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	s := <-c
	fmt.Printf("%+v received, will quit\n", s)
	defer matchingengine.MatchingEngineClose()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("err: %+v\n", err)
	}
	if <-ctx.Done(); true {
		fmt.Printf("timeout of 1 second\n")
	}
	fmt.Printf("Server exiting\n")
}
