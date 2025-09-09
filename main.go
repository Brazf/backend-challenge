package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Backup struct {
	Nome              string    `json:"nome"`
	TamanhoBytes      int64     `json:"tamanho_bytes"`
	DataCriacao       time.Time `json:"data_criacao"`
	UltimaModificacao time.Time `json:"ultima_modificacao"`
}

const (
	sourceDir      = "./valcann/backupsFrom"
	destinationDir = "./valcann/backupsTo"
	logFrom        = "./valcann/backupsFrom.log"
	logTo          = "./valcann/backupsTo.log"
	daysLimit      = 3
)

func main() {
	// Criar pastas se não existirem
	_ = os.MkdirAll(sourceDir, os.ModePerm)
	_ = os.MkdirAll(destinationDir, os.ModePerm)

	// Carregar mock
	mockFile := "./valcann/mock.json"
	backups, err := loadMock(mockFile)
	if err != nil {
		fmt.Println("Erro ao carregar mock:", err)
		return
	}

	// Criar log de origem
	if err := createLog(backups, logFrom); err != nil {
		fmt.Println("Erro ao criar log de origem:", err)
		return
	}

	// Processar arquivos
	var toCopy []Backup
	var toDelete []Backup
	limit := time.Now().AddDate(0, 0, -daysLimit)

	for _, b := range backups {
		if b.DataCriacao.Before(limit) {
			toDelete = append(toDelete, b)
		} else {
			toCopy = append(toCopy, b)
		}
	}

	// Remover arquivos antigos
	for _, f := range toDelete {
		filePath := filepath.Join(sourceDir, f.Nome)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			fmt.Println("Erro ao remover:", f.Nome, err)
		} else {
			fmt.Println("Removido:", f.Nome)
		}
	}

	// Copiar arquivos novos
	for _, f := range toCopy {
		src := filepath.Join(sourceDir, f.Nome)
		dst := filepath.Join(destinationDir, f.Nome)
		if err := copyFile(src, dst); err != nil {
			fmt.Println("Erro ao copiar:", f.Nome, err)
		} else {
			fmt.Println("Copiado:", f.Nome)
		}
	}

	// Criar log de destino
	if err := createLog(toCopy, logTo); err != nil {
		fmt.Println("Erro ao criar log de destino:", err)
		return
	}

	fmt.Println("Processo concluído com sucesso!")
}

func loadMock(path string) ([]Backup, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var backups []Backup
	if err := json.Unmarshal(data, &backups); err != nil {
		return nil, err
	}
	return backups, nil
}

func createLog(backups []Backup, logPath string) error {
	f, err := os.Create(logPath)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, b := range backups {
		line := fmt.Sprintf(
			"Nome: %s | Tamanho: %d bytes | Criado: %s | Última modificação: %s\n",
			b.Nome, b.TamanhoBytes, b.DataCriacao.Format(time.RFC3339), b.UltimaModificacao.Format(time.RFC3339),
		)
		if _, err := f.WriteString(line); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}
