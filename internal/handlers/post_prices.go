package handlers

import (
	"net/http"
)

// HandlerPostPrices обрабатывает POST-запрос для загрузки данных
func HandlerPostPrices() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("Загружаем данные"))
	}
}
