package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

func (poc PortfolioOptimizatiorClient) Recommend(data []entities.StocksHistory) (entities.RecommendationInfoResp, error) {
	logger := poc.logger.With(zap.String("method", "Recommend"))

	start := time.Now()

	var body []byte
	body, err := json.Marshal(map[string]any{
		"data": data,
	})
	if err != nil {
		return entities.RecommendationInfoResp{}, fmt.Errorf("marshall request data: %w", err)
	}
	path, err := url.JoinPath(poc.url, "/recommendation/run")
	if err != nil {
		return entities.RecommendationInfoResp{}, fmt.Errorf("build request url: %w", err)
	}

	logger.Info(fmt.Sprintf("request url: %s, path %v", poc.url, path))
	resp, err := poc.client.Post(path, "application/json", bytes.NewReader(body))
	logger.Info("finish run", zap.Duration("duration", time.Since(start)))
	if err != nil {
		return entities.RecommendationInfoResp{}, fmt.Errorf("send Post requst: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return entities.RecommendationInfoResp{}, fmt.Errorf("responded with %v http code", resp.StatusCode)
	}

	var r entities.RecommendationInfoResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return entities.RecommendationInfoResp{}, fmt.Errorf("decode response: %w", err)
	}

	return r, nil
}
