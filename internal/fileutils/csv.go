package fileutils

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"project-sem/internal/myDB"
)

// CreateCSVFromPrices создает CSV файл из массива цен
func CreateCSVFromPrices(prices []myDB.Price) (*bytes.Buffer, error) {
	var csvBuffer bytes.Buffer
	writer := csv.NewWriter(&csvBuffer)

	// Write CSV header
	if err := writer.Write([]string{"id", "name", "category", "price", "create_date"}); err != nil {
		return nil, fmt.Errorf("CSV header error: %w", err)
	}

	// Process prices
	for _, price := range prices {
		if err := writer.Write([]string{
			price.ID,
			price.Name,
			price.Category,
			fmt.Sprintf("%.2f", price.Price),
			price.CreateDate,
		}); err != nil {
			log.Printf("CSV write error: %v", err)
			continue
		}
	}
	writer.Flush()

	return &csvBuffer, nil
}

// CreateZipFromCSV создает ZIP архив из CSV данных
func CreateZipFromCSV(csvData *bytes.Buffer) (*bytes.Buffer, error) {
	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)
	
	dataFile, err := zipWriter.Create("data.csv")
	if err != nil {
		return nil, fmt.Errorf("ZIP create error: %w", err)
	}

	if _, err := dataFile.Write(csvData.Bytes()); err != nil {
		return nil, fmt.Errorf("ZIP write error: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("ZIP close error: %w", err)
	}

	return &zipBuffer, nil
} 