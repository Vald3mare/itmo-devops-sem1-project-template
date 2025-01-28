package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

// HandlerGetPrices обрабатывает GET-запрос для выгрузки данных
func HandlerGetPrices(db *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		rows, err := db.Query(context.Background(), "SELECT id, name, category, price, create_date FROM prices")
		if err != nil {
			http.Error(w, "не удалось выгрузить данные из БД", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var data [][]string
		data = append(data, []string{"id", "name", "category", "price", "create_date"})

		for rows.Next() {
			var id int
			var name, category string
			var price float64
			var createDate time.Time

			err := rows.Scan(&id, &name, &category, &price, &createDate)
			if err != nil {
				http.Error(w, "не удалось считать данные из строки: "+err.Error(), http.StatusInternalServerError)
				return
			}

			data = append(data, []string{
				strconv.Itoa(id),                     // Преобразуем int в string
				name,                                // name уже string
				category,                            // category уже string
				strconv.FormatFloat(price, 'f', 2, 64), // Преобразуем float64 в string
				createDate.Format("2006-01-02"),     // Преобразуем time.Time в string (формат даты)
			})
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "ошибка итерирования по столбцам БД", http.StatusInternalServerError)
			return
		}

		if len(data) == 1 { // Только заголовок
			http.Error(w, "нету данных в БД", http.StatusNotFound)
			return
		}

		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)
		csvFile, err := zipWriter.Create("data.csv")
		if err != nil {
			http.Error(w, "не удалось создать zip-архив из файла", http.StatusInternalServerError)
			return
		}

		csvWriter := csv.NewWriter(csvFile)
		if err := csvWriter.WriteAll(data); err != nil {
			http.Error(w, "не удалось записать данные csv в архив", http.StatusInternalServerError)
			return
		}
		csvWriter.Flush()

		if err := zipWriter.Close(); err != nil {
			http.Error(w, "ошибка при закрытии архива", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(buf.Bytes()); err != nil {
			http.Error(w, "ошибка в отправке response", http.StatusInternalServerError)
			return
		}
	}
}