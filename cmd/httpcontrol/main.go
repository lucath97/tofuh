package main

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"lucathurm.dev/tofuh/internal/auth"
	"lucathurm.dev/tofuh/internal/config"
	"lucathurm.dev/tofuh/internal/db"
)

type Response struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func parseOffset(offsetParam string) (uint8, error) {
	offset, parseErr := strconv.ParseUint(offsetParam, 10, 8)
	if parseErr != nil {
		return 0, parseErr
	}
	return uint8(offset), nil
}

func getStateHandler(ctx *gin.Context, log *slog.Logger, rds *redis.Client, cfg config.Config) {
	log.Info("received request to get state")

	state, getErr := db.GetState(rds, ctx, cfg.DbStateKey)
	if getErr != nil {
		log.Error("failed to get state from db", "error", getErr)
		ctx.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}
	log.Info("got state from db")

	ctx.JSON(http.StatusOK, Response{Message: "OK", Data: state})
}

func setBitHandler(ctx *gin.Context, log *slog.Logger, rds *redis.Client, cfg config.Config) {
	log.Info("received request to set bit")

	offsetParam := ctx.Param("offset")
	offset, parseErr := parseOffset(ctx.Param("offset"))
	if parseErr != nil {
		log.Error("failed to parse bit offset", "error", parseErr, "offset", offsetParam)
		ctx.JSON(http.StatusBadRequest, Response{Message: "offset must be a uint8"})
		return
	}

	setErr := db.SetBit(rds, ctx, cfg.DbStateKey, offset, true)
	if setErr != nil {
		log.Error("failed to set bit in db", "error", setErr)
		ctx.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}
	log.Info("set bit in db")

	ctx.JSON(http.StatusOK, Response{Message: "OK"})
}

func unsetBitHandler(ctx *gin.Context, log *slog.Logger, rds *redis.Client, cfg config.Config) {
	log.Info("received request to unset bit")

	offsetParam := ctx.Param("offset")
	offset, parseErr := parseOffset(ctx.Param("offset"))
	if parseErr != nil {
		log.Error("failed to parse bit offset", "error", parseErr, "offset", offsetParam)
		ctx.JSON(http.StatusBadRequest, Response{Message: "offset must be a uint8"})
		return
	}

	setErr := db.SetBit(rds, ctx, cfg.DbStateKey, offset, false)
	if setErr != nil {
		log.Error("failed to unset bit in db", "error", setErr)
		ctx.JSON(http.StatusInternalServerError, Response{Message: "Internal Server Error"})
		return
	}
	log.Info("unset bit in db")

	ctx.JSON(http.StatusOK, Response{Message: "OK"})
}

func authMiddleware(log *slog.Logger, authService *auth.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.Info("initialized auth")

		apiKey := ctx.GetHeader("Authorization")
		client, authErr := authService.CheckAPIKey(apiKey)
		if authErr != nil {
			log.Warn("client is not authenticated")
			ctx.JSON(http.StatusUnauthorized, Response{Message: "Unauthorized"})
			ctx.Abort()
		}

		log.Info("authenticated client", "clientId", client.Id)
		ctx.Next()
	}
}

func main() {
	log := slog.Default()

	cfg := config.LoadConfig()
	log.Info("loaded config")

	authService, authServiceErr := auth.NewAuthService(cfg.HTTPClientsFile)
	if authServiceErr != nil {
		log.Error("failed to initialize auth service", "error", authServiceErr)
		panic(authServiceErr)
	}

	rds := redis.NewClient(&redis.Options{Addr: cfg.DbAddress, Password: cfg.DbPassword})
	log.Info("created redis client")

	initCtx, initCtxCancel := context.WithTimeout(context.Background(), cfg.DbTimeout)
	if initErr := db.InitState(rds, initCtx, cfg.DbStateKey); initErr != nil {
		log.Error("failed to connect to database", "error", initErr)
		initCtxCancel()
		panic(initErr)
	}
	initCtxCancel()
	log.Info("connected to database")

	app := gin.Default()

	app.SetTrustedProxies(nil)

	grpAuthenticated := app.Group("/api")
	grpAuthenticated.Use(authMiddleware(log, &authService))

	app.GET("/api/state", func(ctx *gin.Context) {
		getStateHandler(ctx, log, rds, cfg)
	})
	app.POST("/api/state/:offset", func(ctx *gin.Context) {
		setBitHandler(ctx, log, rds, cfg)
	})
	app.DELETE("/api/state/:offset", func(ctx *gin.Context) {
		unsetBitHandler(ctx, log, rds, cfg)
	})

	runErr := app.Run(cfg.HTTPAddress)
	if runErr != nil {
		slog.Error("failed to start http server", "error", runErr)
		panic(runErr)
	}
}
