package handlers

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"project-sem/internal/myDB"
)

// HandlerGetPrices обрабатывает GET-запрос для получения данных из базы данных
func HandlerPostPrices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Error parsing multipart form: "+err.Error(), http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		tempFile, err := os.CreateTemp("", "upload-*.zip")
		if err != nil {
			http.Error(w, "Error creating temp file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		if _, err = io.Copy(tempFile, file); err != nil {
			http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		zipReader, err := zip.OpenReader(tempFile.Name())
		if err != nil {
			http.Error(w, "Error opening zip: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer zipReader.Close()

		var csvFile *zip.File
		for _, f := range zipReader.File {
			if f.Name == "data.csv" {
				csvFile = f
				break
			}
		}
		if csvFile == nil {
			http.Error(w, "data.csv not found in archive", http.StatusBadRequest)
			return
		}

		rc, err := csvFile.Open()
		if err != nil {
			http.Error(w, "Error opening csv: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rc.Close()

		reader := csv.NewReader(rc)
		records, err := reader.ReadAll()
		if err != nil {
			http.Error(w, "Error reading csv: "+err.Error(), http.StatusInternalServerError)
			return
		}

		totalItems, totalCategories, totalPrice, err := myDB.InsertPrices(records)
		if err != nil {
			http.Error(w, "Error inserting data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"total_items":      totalItems,
			"total_categories": totalCategories,
			"total_price":      totalPrice,
		}

		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
