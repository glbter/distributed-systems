package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	"go.uber.org/zap"

	"github.com/glbter/distributed-systems/main-service/entities"
)

type PortfolioOptimizatiorClient struct {
	url    string
	client *http.Client
	logger *zap.Logger
}

func NewClient(c *http.Client, url string, logger *zap.Logger) PortfolioOptimizatiorClient {
	return PortfolioOptimizatiorClient{
		url:    url,
		client: c,
		logger: logger.With(zap.String("caller", "PortfolioOptimizatiorClient")),
	}
}

func (poc PortfolioOptimizatiorClient) Recommend(data []entities.StocksHistory) (entities.RecommendedPortfolio, error) {
	logger := poc.logger.With(zap.String("method", "Recommend"))

	start := time.Now()

	var body []byte
	body, err := json.Marshal(data)
	if err != nil {
		return entities.RecommendedPortfolio{}, err
	}

	resp, err := poc.client.Post(path.Join(poc.url+"/run"), "application/json", bytes.NewReader(body))
	logger.Debug("finish run", zap.Duration("duration", time.Since(start)))
	if err != nil {
		return entities.RecommendedPortfolio{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return entities.RecommendedPortfolio{}, fmt.Errorf("responded with %v http code", resp.StatusCode)
	}

	var r recommendationResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return entities.RecommendedPortfolio{}, err
	}

	return entities.RecommendedPortfolio{}, nil
}

type recommendationResp struct {
	Returns   float64                                       `json:"returns"`
	Risk      float64                                       `json:"risk"`
	Portfolio map[StockTicker]entities.RecommendedPortfolio `json:"portfolio"`
}

type StockTicker string

//type entities.RecommendedPortfolio struct {
//	Mon3  float64 `json:"3_mon_return"`
//	Mon6  float64 `json:"6_mon_return"`
//	Mon12 float64 `json:"12_mon_return"`
//	Mon24 float64 `json:"24_mon_return"`
//	Mon32 float64 `json:"32_mon_return"`
//}
