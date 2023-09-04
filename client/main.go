package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)

	if err != nil {
		fmt.Println("Erro ao criar a solicitação:", err)
		return
	}

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Erro ao fazer a solicitação:", err)
		return
	}
	defer res.Body.Close()

	fmt.Println("Status da resposta:", res.Status)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Erro ao ler o corpo da resposta:", err)
		return
	}

	fmt.Println("Corpo da resposta:", string(body))

	var quotation string
	err = json.Unmarshal([]byte(body), &quotation)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return
	}

	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar arquivo: %v\n", err)
	}

	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("Dólar: %s", quotation))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar arquivo: %v\n", err)
	}
}
