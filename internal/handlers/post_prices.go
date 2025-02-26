package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HandlerPostPrices обрабатывает POST-запрос для загрузки цен
func HandlerPostPrices(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20) // 10 MB
		if err != nil {
			http.Error(w, "Неудалось извлечь вложения запроса", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Не удалось извлечь файл из вложений запроса", http.StatusBadRequest)
			return
		}
		defer file.Close()

		fmt.Printf("Received file: %s\n", header.Filename)

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Не удалось прочитать файл", http.StatusInternalServerError)
			return
		}

		zipReader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
		if err != nil {
			http.Error(w, "Не удалось разархивировать файл запроса", http.StatusInternalServerError)
			return
		}

		for _, zipFile := range zipReader.File {
			if !strings.HasSuffix(zipFile.Name, ".csv") {
				continue
			}
			fmt.Printf("Processing CSV file: %s\n", zipFile.Name)

			fileInZip, err := zipFile.Open()
			if err != nil {
				http.Error(w, fmt.Sprintf("Не удалось открыть файл %s внутри zip-архива", zipFile.Name), http.StatusInternalServerError)
				return
			}
			defer fileInZip.Close()

			csvReader := csv.NewReader(fileInZip)
			records, err := csvReader.ReadAll()
			if err != nil {
				http.Error(w, fmt.Sprintf("Не удалось прочитать CSV файл %s", zipFile.Name), http.StatusInternalServerError)
				return
			}

			for i, record := range records {
				if i == 0 {
					continue
				}

				productID, _ := strconv.Atoi(record[0])
				productName := record[1]
				category := record[2]
				price, _ := strconv.ParseFloat(record[3], 64)
				createdAt := record[4]

				_, err := pool.Exec(context.Background(),
					"INSERT INTO prices (product_id, product_name, category, price, created_at) VALUES ($1, $2, $3, $4, $5)",
					productID, productName, category, price, createdAt,
				)
				if err != nil {
					fmt.Printf("Не удалось разместить запись в БД %d: %v\n", i, err)
					continue
				}

				fmt.Printf("Запись размещена в БД: %v\n", record)
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Файл обработан успешно"))
	}
}
