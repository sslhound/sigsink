package config

import (
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

var ProductionEnvironment = "production"

func Logger(c *cli.Context) (*zap.Logger, error) {
	if c.String("environment") == ProductionEnvironment {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

func ListenAddress(cliCtx *cli.Context) string {
	if listen := cliCtx.String("listen"); len(listen) > 0 {
		return listen
	}
	return ":8080"
}
