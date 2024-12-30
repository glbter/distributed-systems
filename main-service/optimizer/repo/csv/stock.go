package csv

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/glbter/distributed-systems/main-service/entities"
)

type StockRepo struct {
}

func (StockRepo) GetStocks() ([]entities.StocksHistory, error) {
	stocks := make([]entities.StocksHistory, 0, 6)

	files := []string{"hdfc", "itc", "l&t", "m&m", "sunpha", "tcs"}

	for _, file := range files {
		content, err := readCsvFile("optimizer/repo/csv/Stocks_Data/" + file + ".csv")
		if err != nil {
			return nil, err
		}

		stockData := make([]entities.StockDateData, 0)
		for _, line := range content[1:] {
			cl, err := strconv.ParseFloat(line[1], 64)
			if err != nil {
				return nil, err
			}

			stockData = append(stockData, entities.StockDateData{
				Date:  line[0],
				Close: cl,
			})
		}

		stocks = append(stocks, entities.StocksHistory{
			Name: file,
			Data: stockData,
		})

	}

	return stocks, nil
}

func readCsvFile(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}
