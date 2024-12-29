package main

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"net/http"
	"os"
	"time"

	optimizerHttp "distributed-systems/main-service/optimizer/http"
	"distributed-systems/main-service/optimizer/repo/csv"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Create a custom encoder configuration
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Apply the custom encoder configuration using WithOptions
	customLogger := logger.WithOptions(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig), // Custom encoder
			zapcore.AddSync(os.Stdout),            // Sync to stdout
			c,                                     // Use the same level as the original core
		)
	}))

	client := &http.Client{Timeout: time.Second * 60}

	url := os.Getenv("URL_PORTFOLIO_ENGINE")
	if url == "" {
		log.Fatalln("client url is empty")
	}

	handler := PortfolioHandler{
		PortfolioEngine: optimizerHttp.NewClient(client, url, customLogger),
		StockRepo:       csv.StockRepo{},
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Post("portfolio/optimize", handler.OptimizePortfolio)

	http.ListenAndServe(":8080", r)
}

type PortfolioHandler struct {
	StockRepo       csv.StockRepo
	PortfolioEngine optimizerHttp.PortfolioOptimizatiorClient
}

func (h PortfolioHandler) OptimizePortfolio(w http.ResponseWriter, _ *http.Request) {
	stocks, err := h.StockRepo.GetStocks()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	res, err := h.PortfolioEngine.Recommend(stocks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}
