package http

import (
	"encoding/json"
	"fmt"
	optimizerHttp "github.com/glbter/distributed-systems/main-service/optimizer/client/http"
	"github.com/glbter/distributed-systems/main-service/optimizer/repo/csv"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"net/http"
	"strconv"

	optimizerRabbit "github.com/glbter/distributed-systems/main-service/optimizer/client/rabbit"
	"github.com/google/uuid"
)

type PortfolioHandler struct {
	Logger               *zap.Logger
	StockRepo            csv.StockRepo
	PortfolioEngine      *optimizerHttp.PortfolioOptimizatiorClient
	PortfolioEngineAsync *optimizerRabbit.PortfolioOptimizatiorClient
	ReplyMessageCh       <-chan amqp.Delivery
}

func (h PortfolioHandler) OptimizePortfolio(w http.ResponseWriter, _ *http.Request) {
	logger := h.Logger.With(zap.String("method", "OptimizePortfolio"))

	stocks, err := h.StockRepo.GetStocks()
	if err != nil {
		logger.Error(fmt.Errorf("get stocks: %w", err).Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := h.PortfolioEngine.Recommend(stocks)
	if err != nil {
		logger.Error(fmt.Errorf("recommend portfolio: %w", err).Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		logger.Error(fmt.Errorf("encode response: %w", err).Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h PortfolioHandler) OptimizePortfolioAsync(w http.ResponseWriter, r *http.Request) {
	logger := h.Logger.With(zap.String("method", "OptimizePortfolio"))

	stocks, err := h.StockRepo.GetStocks()
	if err != nil {
		logger.Error(fmt.Errorf("get stocks: %w", err).Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vip := r.URL.Query().Get("is_vip")
	isVip, err := strconv.ParseBool(vip)
	if err != nil {
		logger.Error(fmt.Errorf("parse is_vip: %w", err).Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cid := uuid.New().String()
	err = h.PortfolioEngineAsync.StartRecommend(r.Context(), stocks, cid, isVip)
	if err != nil {
		logger.Error(fmt.Errorf("recommend portfolio: %w", err).Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for d := range h.ReplyMessageCh {
		if cid == d.CorrelationId {
			logger.Info(fmt.Sprintf("received a message: %s", d.Body))
		}

		w.Write(d.Body)
		return
	}
}
