package privat_bank

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	"go.uber.org/zap"
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

type StocksHistory struct {
}

type RecommendedPortfolio struct {
}

func (poc PortfolioOptimizatiorClient) Recommed(data StocksHistory) (RecommendedPortfolio, error) {
	logger := poc.logger.With(zap.String("method", "Recommend"))

	start := time.Now()
	resp, err := poc.client.Get(path.Join(poc.url + "/run"))
	logger.Debug("finish run", zap.Duration("duration", time.Since(start)))
	if err != nil {
		return RecommendedPortfolio{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return RecommendedPortfolio{}, fmt.Errorf("responded with %v http code", resp.StatusCode)
	}

	var r recommendationResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return RecommendedPortfolio{}, err
	}

	return RecommendedPortfolio{}, nil
}

type recommendationResp struct {
}
