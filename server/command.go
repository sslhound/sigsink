package server

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/oklog/run"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/sslhound/sigsink/config"
)

func Action(cliCtx *cli.Context) error {
	logger, err := config.Logger(cliCtx)
	if err != nil {
		return err
	}

	domain := cliCtx.String("domain")
	siteBase := fmt.Sprintf("https://%s", domain)

	logger.Info("Starting",
		zap.String("GOOS", runtime.GOOS),
		zap.String("site", siteBase),
		zap.String("env", cliCtx.String("environment")))

	r := gin.New()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))

	k := &keyCache{
		keys:              make(map[string]*rsa.PublicKey),
		enableFetchRemote: cliCtx.Bool("enable-keyfetch"),
		logger:            logger,
	}

	for _, path := range cliCtx.StringSlice("key-source") {
		if err = k.addDirectory(path); err != nil {
			return err
		}
	}

	h := handler{keyCache: k}

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r.NoRoute(h.verifySignatureHandler)

	var group run.Group

	srv := &http.Server{
		Addr:    config.ListenAddress(cliCtx),
		Handler: r,
	}

	group.Add(func() error {
		logger.Info("starting http service", zap.String("addr", srv.Addr))
		return srv.ListenAndServe()
	}, func(error) {
		httpCancelCtx, httpCancelCtxCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer httpCancelCtxCancel()
		logger.Info("stopping http service")
		if err = srv.Shutdown(httpCancelCtx); err != nil {
			logger.Error("error stopping http service", zap.Error(err))
		}
	})

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	group.Add(func() error {
		logger.Info("starting signal listener")
		<-quit
		return nil
	}, func(error) {
		logger.Info("stopping signal listener")
		close(quit)
	})

	return group.Run()
}
