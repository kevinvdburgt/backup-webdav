package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: backup-webdav <projectname> <filename>")
		os.Exit(1)
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	project := os.Args[1]
	filename := os.Args[2]

	remoteURL := fmt.Sprintf("%s/%s-%s", project, time.Now().Format("2006/01-02/150405"), filename)

	baseURL := env("BASE_URL")
	username := env("USERNAME")
	password := env("PASSWORD")
	publicKey := env("PUBLIC_KEY")

	publicKeyData := []byte(publicKey)
	recipient, err := loadRecipient(publicKeyData)
	if err != nil {
		log.Fatalf("Failed to load GPG key: %v", err)
	}

	err = mkdirs(baseURL, username, password, filepath.Dir(remoteURL))
	if err != nil {
		log.Fatalf("Error creating directories: %v", err)
	}

	// Setup pipeline: stdin -> gzip -> pgp encrypt -> http
	reader, writer := io.Pipe()
	go func() {
		gzipWriter := gzip.NewWriter(writer)
		pgpWriter, err := openpgp.Encrypt(gzipWriter, []*openpgp.Entity{recipient}, nil, nil, &packet.Config{})
		if err != nil {
			writer.CloseWithError(err)
			return
		}

		if _, err := io.Copy(pgpWriter, os.Stdin); err != nil {
			writer.CloseWithError(err)
			return
		}

		pgpWriter.Close()
		gzipWriter.Close()
		writer.Close()
	}()

	req, err := http.NewRequest("PUT", baseURL+"/"+remoteURL, reader)
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %v", err)
	}

	req.SetBasicAuth(username, password)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		log.Fatalf("Upload failed with status: %s", resp.Status)
	}

	log.Printf("File uploaded successfully to %s/%s", baseURL, remoteURL)
}

func env(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Missing required environment variable: %s", key)
	}
	return val
}

func loadRecipient(publicKey []byte) (*openpgp.Entity, error) {
	keyReader := bytes.NewReader(publicKey)
	entities, err := openpgp.ReadArmoredKeyRing(keyReader)
	if err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		return nil, fmt.Errorf("no keys found in keyring")
	}

	return entities[0], nil
}

func mkdirs(baseURL, username, password, path string) error {
	parts := strings.Split(path, "/")
	current := baseURL
	client := &http.Client{}

	for _, part := range parts {
		if part == "" {
			continue
		}

		current += "/" + part
		req, _ := http.NewRequest("MKCOL", current, nil)
		req.SetBasicAuth(username, password)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("MKCOL failed: %w", err)
		}

		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to create dir %s (status %d)", current, resp.StatusCode)
		}
	}

	return nil
}
