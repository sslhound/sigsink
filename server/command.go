package server

import (
	"context"
	"crypto/rsa"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/oklog/run"
	"github.com/urfave/cli/v2"
)

func Action(cliCtx *cli.Context) error {
	k := &keyCache{
		keys:              make(map[string]*rsa.PublicKey),
		enableFetchRemote: cliCtx.Bool("enable-keyfetch"),
	}

	for _, path := range cliCtx.StringSlice("key-source") {
		if err := k.addDirectory(path); err != nil {
			return err
		}
	}

	h := handler{keyCache: k}

	mux := http.NewServeMux()

	mux.HandleFunc("/", h.ServeHTTP)

	var group run.Group

	listen := ":7000"
	if listenArg := cliCtx.String("listen"); len(listenArg) > 0 {
		listen = listenArg
	}

	srv := &http.Server{
		Addr:    listen,
		Handler: mux,
	}

	group.Add(func() error {
		log.Println("starting http service", srv.Addr)
		return srv.ListenAndServe()
	}, func(error) {
		httpCancelCtx, httpCancelCtxCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer httpCancelCtxCancel()
		log.Println("stopping http service")
		if err := srv.Shutdown(httpCancelCtx); err != nil {
			log.Println("error stopping http service:", err.Error())
		}
	})

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	group.Add(func() error {
		log.Println("starting signal listener")
		<-quit
		return nil
	}, func(error) {
		log.Println("stopping signal listener")
		close(quit)
	})

	return group.Run()
}
