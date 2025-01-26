package handlers

import (
	"archive/zip"
	"encoding/csv"
	"net/http"
	"os"
	"project/db"
	"strconv"

	"context"
)

func GetPrices(w http.ResponseWriter, r *http.Request) {
	// Создание временного файла
	tempFile, err := os.CreateTemp("", "data-*.zip")
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	// Создание ZIP-архива
	zipWriter := zip.NewWriter(tempFile)
	fileWriter, err := zipWriter.Create("data.csv")
	if err != nil {
		http.Error(w, "Failed to create CSV in zip", http.StatusInternalServerError)
		return
	}

	// Пишем данные в CSV
	writer := csv.NewWriter(fileWriter)
	query := "SELECT id, name, category, price, create_date FROM prices"
	rows, err := db.DBPool.Query(context.Background(), query)
	if err != nil {
		http.Error(w, "Failed to fetch data from database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, category, createDate string
		var price float64

		err := rows.Scan(&id, &name, &category, &price, &createDate)
		if err != nil {
			http.Error(w, "Failed to scan row", http.StatusInternalServerError)
			return
		}

		writer.Write([]string{strconv.Itoa(id), name, category, strconv.FormatFloat(price, 'f', 2, 64), createDate})
	}

	writer.Flush()
	zipWriter.Close()

	// Отправляем файл клиенту
	tempFile.Seek(0, 0)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="data.zip"`)
	http.ServeFile(w, r, tempFile.Name())
}
