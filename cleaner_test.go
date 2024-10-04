package main

/*
To generate test data:

<testdata/web.log go run cleaner.go \
  -crawler-output ./testdata/expected-crawler.log \
  -non-crawler-output ./testdata/expected-non-crawler.log \
  -error-output ./testdata/expected-error.log \
  -extra-crawler-agents-file testdata/extra-crawler-agents.txt
*/

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func readFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return data
}

var extraCrawlerAgentsBytes []byte = readFile("testdata/extra-crawler-agents.txt")
var rawLogBytes []byte = readFile("testdata/web.log")
var expectedNonCrawlerBytes []byte = readFile("testdata/expected-non-crawler.log")
var expectedCrawlerBytes []byte = readFile("testdata/expected-crawler.log")
var expectedErrorBytes []byte = readFile("testdata/expected-error.log")

func TestCleaner(t *testing.T) {
	var nonCrawlerBytes, crawlerBytes, errorBytes bytes.Buffer

	addExtraCrawlerAgents(bytes.NewReader(extraCrawlerAgentsBytes))

	cleanCrawlers(defaultUserAgentKey, bytes.NewReader(rawLogBytes),
		io.Writer(&nonCrawlerBytes), io.Writer(&crawlerBytes), io.Writer(&errorBytes))

	if !bytes.Equal(expectedNonCrawlerBytes, nonCrawlerBytes.Bytes()) {
		t.Fatal("Non crawler logs did not match expected")
	}

	if !bytes.Equal(expectedCrawlerBytes, crawlerBytes.Bytes()) {
		t.Fatal("Crawler logs did not match expected")
	}

	if !bytes.Equal(expectedErrorBytes, errorBytes.Bytes()) {
		t.Fatal("Error logs did not match expected")
	}
}
