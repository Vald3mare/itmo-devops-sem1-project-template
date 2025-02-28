package main

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	dbURL = "postgres://validator:val1dat0r@localhost:5432/project-sem-1?sslmode=disable"
)

func main() {
	r := gin.Default()

	r.POST("/api/v0/prices", handlePost)
	r.GET("/api/v0/prices", handleGet)

	r.Run(":8080")
}

func handlePost(c *gin.Context) {
	// Обработка загрузки файла
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	// Распаковка ZIP
	reader, err := zip.NewReader(file, fileHeader.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var csvFile *zip.File
	for _, f := range reader.File {
		if f.Name == "data.csv" {
			csvFile = f
			break
		}
	}

	if csvFile == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data.csv not found"})
		return
	}

	rc, err := csvFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rc.Close()

	// Подключение к БД
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stmt, err := tx.Prepare("INSERT INTO prices (product_id, created_at, product_name, category, price) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()

	// Чтение CSV
	csvReader := csv.NewReader(rc)
	totalItems := 0
	totalPrice := 0.0
	categories := make(map[string]struct{})

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		productID, _ := strconv.Atoi(record[0])
		createdAt, _ := time.Parse("2006-01-02", record[1])
		productName := record[2]
		category := record[3]
		price, _ := strconv.ParseFloat(record[4], 64)

		_, err = stmt.Exec(productID, createdAt, productName, category, price)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		totalItems++
		totalPrice += price
		categories[category] = struct{}{}
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_items":      totalItems,
		"total_categories": len(categories),
		"total_price":      totalPrice,
	})
}

func handleGet(c *gin.Context) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM prices")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Создание временного CSV
	tmpFile, err := os.CreateTemp("", "data-*.csv")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	writer := csv.NewWriter(tmpFile)
	defer writer.Flush()

	for rows.Next() {
		var (
			productID   int
			createdAt   time.Time
			productName string
			category    string
			price       float64
		)

		err = rows.Scan(&productID, &createdAt, &productName, &category, &price)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		writer.Write([]string{
			strconv.Itoa(productID),
			createdAt.Format("2006-01-02"),
			productName,
			category,
			strconv.FormatFloat(price, 'f', -1, 64),
		})
	}

	writer.Flush()
	tmpFile.Close()

	// Создание ZIP
	zipFile, err := os.CreateTemp("", "data-*.zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer os.Remove(zipFile.Name())
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	f, err := zipWriter.Create("data.csv")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	content, _ := os.ReadFile(tmpFile.Name())
	f.Write(content)

	c.FileAttachment(zipFile.Name(), "data.zip")
}
