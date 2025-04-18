package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/mmcdole/gofeed"
)

func stdinToArgs() ([]string, error) {
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("error parsing stdin: %s", err)
		}
		return strings.Fields(string(input)), nil
	}
	return []string{}, nil
}

func fileToArgs(filepath string) ([]string, error) {
	if filepath == "" {
		return []string{}, nil
	}
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening input file: %s", err)
	}
	defer file.Close()
	var args []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		input := scanner.Text()
		if input != "" {
			args = append(args, input)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning input file: %s", err)
	}
	return args, nil
}

func domain(item *gofeed.Item) string {
	url, err := url.Parse(item.Link)
	if err != nil {
		log.Printf("WARNING: fail to parse domain %s: %s\n", item.Link, err)
	}
	return strings.TrimPrefix(url.Hostname(), "www.")
}

func publication(item *gofeed.Item) string {
	if item.PublishedParsed == nil {
		if item.Published != "" {
			return item.Published
		}
		return "Once upon a time"
	}
	return item.PublishedParsed.Format("2006-01-02")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func generateNonce(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil { // Very unlikely
		log.Fatal(fmt.Errorf("failed to generate nonce: %w", err))
	}
	return base64.URLEncoding.EncodeToString(b)
}
