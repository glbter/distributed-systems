package cmd

import (
	portfolioHttp "github.com/glbter/distributed-systems/main-service/http"
	optimizerHttp "github.com/glbter/distributed-systems/main-service/optimizer/client/http"
	"github.com/glbter/distributed-systems/main-service/optimizer/repo/csv"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"time"
)

func ExecuteSync() {
	logger := InitLogger()

	client := &http.Client{Timeout: time.Second * 60}

	url := os.Getenv("URL_PORTFOLIO_ENGINE")
	if url == "" {
		log.Fatalln("client url is empty")
	}

	httpClient := optimizerHttp.NewClient(client, url, logger)
	handler := portfolioHttp.PortfolioHandler{
		PortfolioEngine: &httpClient,
		StockRepo:       csv.StockRepo{},
		Logger:          logger,
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Post("/portfolio/optimize", handler.OptimizePortfolio)

	http.ListenAndServe(":8080", r)
}
