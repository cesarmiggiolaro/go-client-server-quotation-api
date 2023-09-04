package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

const file string = "./quotation.db"

type Quotation struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", GetQuotationHandler)
	http.ListenAndServe(":8080", mux)
}

func GetQuotationHandler(w http.ResponseWriter, r *http.Request) {

	quotation, err := GetExternalQuotation()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(quotation.Usdbrl.Bid)

	saveQuotation(quotation)
}

func GetExternalQuotation() (*Quotation, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)

	if err != nil {
		fmt.Println("Erro ao criar a solicitação:", err)
		return nil, err
	}

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Erro ao fazer a solicitação:", err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Erro ao ler o corpo da resposta:", err)
		return nil, err
	}

	var data Quotation
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("Erro ao fazer o parse da resposta:", err)
	}

	return &data, nil
}

func saveQuotation(quotation *Quotation) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	db, err := DatabaseConextion()

	if err != nil {
		fmt.Println("Erro ao iniciar database", err)
		return err
	}
	defer db.Close()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		fmt.Println("Erro ao iniciar transação:", err)
		return err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO quotation (id, code, codein, bid, created_at) VALUES (?, ?, ?, ?, ?)",
		uuid.New().String(),
		quotation.Usdbrl.Code,
		quotation.Usdbrl.Codein,
		quotation.Usdbrl.Bid,
		time.Now(),
	)

	if err != nil {
		tx.Rollback()
		fmt.Println("Erro ao inserir registro:", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		fmt.Println("Erro ao fazer commit:", err)
		return err
	}

	fmt.Println("Registro inserido com sucesso")

	return nil
}

func DatabaseConextion() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		fmt.Println("Erro ao abrir conexão:", err)
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS quotation (id VARCHAR(255) PRIMARY KEY, code VARCHAR(3), codein VARCHAR(3), bid DECIMAL(10,4), created_at DATE_TIME)")
	if err != nil {
		fmt.Println("Erro ao criar tabela:", err)
		return nil, err
	}

	fmt.Println("Conexão com o banco de dados SQLite3 estabelecida.")
	return db, nil
}
