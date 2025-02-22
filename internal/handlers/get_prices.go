package handlers

import (
	"net/http"
)

// HandlerGetPrices обрабатывает GET-запрос для выгрузки данных
func HandlerGetPrices() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("Выгружаем данные"))
	}
}
