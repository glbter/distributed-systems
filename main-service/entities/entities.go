package entities

type StocksHistory struct {
	Name string
	Data []StockDateData
}

type StockDateData struct {
	Date  string
	Close float64
}

type RecommendedPortfolio struct {
	Mon3  float64 `json:"3_mon_return"`
	Mon6  float64 `json:"6_mon_return"`
	Mon12 float64 `json:"12_mon_return"`
	Mon24 float64 `json:"24_mon_return"`
	Mon32 float64 `json:"32_mon_return"`
}
