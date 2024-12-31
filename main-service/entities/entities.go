package entities

type StocksHistory struct {
	Name string `json:"name"`
	Data []StockDateData `json:"data"`
}

type StockDateData struct {
	Date  string `json:"date"`
	Close float64 `json:"close"`
}

type RecommendedPortfolio struct {
	Mon3  float64 `json:"3_mon_return"`
	Mon6  float64 `json:"6_mon_return"`
	Mon12 float64 `json:"12_mon_return"`
	Mon24 float64 `json:"24_mon_return"`
	Mon32 float64 `json:"32_mon_return"`
}

type StockTicker string

type RecommendationInfoResp struct {
	Returns   float64                                       `json:"returns"`
	Risk      float64                                       `json:"risk"`
	Portfolio map[StockTicker]RecommendedPortfolio `json:"portfolio"`
}
