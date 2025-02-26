package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func HandlerPostPrices(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error reading file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error reading file content", http.StatusInternalServerError)
			return
		}

		zipReader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
		if err != nil {
			http.Error(w, "Error reading zip file", http.StatusBadRequest)
			return
		}

		var totalItems int
		var totalPrice float64
		categories := make(map[string]struct{})

		for _, f := range zipReader.File {
			if !strings.HasSuffix(f.Name, ".csv") {
				continue
			}

			rc, err := f.Open()
			if err != nil {
				http.Error(w, "Error opening CSV file", http.StatusInternalServerError)
				return
			}

			reader := csv.NewReader(rc)
			records, err := reader.ReadAll()
			if err != nil {
				http.Error(w, "Error reading CSV data", http.StatusInternalServerError)
				return
			}

			for i, record := range records {
				if i == 0 {
					continue
				}

				productID, _ := strconv.Atoi(record[0])
				createdAt := record[1]
				productName := record[2]
				category := record[3]
				price, _ := strconv.ParseFloat(record[4], 64)

				_, err = pool.Exec(context.Background(),
					`INSERT INTO prices (product_id, created_at, product_name, category, price)
					 VALUES ($1, $2, $3, $4, $5)`,
					productID, createdAt, productName, category, price,
				)

				if err != nil {
					fmt.Printf("Error inserting record: %v\n", err)
					continue
				}

				totalItems++
				totalPrice += price
				categories[category] = struct{}{}
			}
			rc.Close()
		}

		response := map[string]interface{}{
			"total_items":      totalItems,
			"total_categories": len(categories),
			"total_price":      int(totalPrice),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
