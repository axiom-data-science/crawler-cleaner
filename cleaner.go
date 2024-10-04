package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/monperrus/crawler-user-agents"
)

var (
	userAgentKeyConfig     string
	extraCrawlerAgentsFile string
	nonCrawlerOutput       string
	crawlerOutput          string
	errorOutput            string
)

//go:embed agent-allowlist.txt
var agentAllowListBytes []byte

var allowedAgentOverrides = func() []string {
	var allowedAgentOverrides []string
	s := bufio.NewScanner(bytes.NewReader(agentAllowListBytes))
	s.Split(bufio.ScanLines)
	for s.Scan() {
		allowedAgent := strings.TrimSpace(s.Text())
		if len(allowedAgent) == 0 {
			continue
		}
		allowedAgentOverrides = append(allowedAgentOverrides, allowedAgent)
	}
	return allowedAgentOverrides
}()

// adapted from crawler-user-agents/validate.go
var crawlerRegexps = func() []*regexp.Regexp {
	regexps := make([]*regexp.Regexp, 0, len(agents.Crawlers))
	for _, crawler := range agents.Crawlers {
		if !slices.Contains(allowedAgentOverrides, crawler.Pattern) {
			regexps = append(regexps, regexp.MustCompile(crawler.Pattern))
		}
	}
	return regexps
}()

// adapted from crawler-user-agents/validate.go
// Returns if User Agent string matches any of crawler patterns.
var isCrawler = func(userAgent string) bool {
	for _, re := range crawlerRegexps {
		if re.MatchString(userAgent) {
			return true
		}
	}
	return false
}

const defaultExtraCrawlerAgentsFile = "extra-crawler-agents.txt"
const defaultUserAgentKey = "http_user_agent"

func getWriter(w string) *os.File {
	switch w {
	case "-", "/dev/stdout", "stdout":
		//stdout
		return os.Stdout
	case "+", "/dev/stderr", "stderr":
		//stderr
		return os.Stderr
	case "0", "/dev/null", "null":
		//devnull
		return os.NewFile(0, os.DevNull)
	default:
		//file path
		file, err := os.OpenFile(w, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		return file
	}
}

func addExtraCrawlerAgents(extraAgentsReader io.Reader) {
	extraAgentsScanner := bufio.NewScanner(extraAgentsReader)
	extraAgentsScanner.Split(bufio.ScanLines)

	for extraAgentsScanner.Scan() {
		extraAgent := strings.TrimSpace(extraAgentsScanner.Text())
		if len(extraAgent) == 0 {
			continue
		}
		crawlerRegexps = append(crawlerRegexps, regexp.MustCompile(extraAgent))
	}
}

func cleanCrawlers(userAgentKey string, logReader io.Reader, nonCrawlerWriter io.Writer,
	crawlerWriter io.Writer, errorWriter io.Writer) {
	s := bufio.NewScanner(logReader)
	for s.Scan() {
		var v map[string]interface{}
		if err := json.Unmarshal(s.Bytes(), &v); err != nil {
			// json parse error
			fmt.Fprintln(errorWriter, s.Text())
			continue
		}

		agent, ok := v[userAgentKey]
		// assume its ok if we don't have agent info
		if ok && isCrawler(agent.(string)) {
			// crawler detected
			fmt.Fprintln(crawlerWriter, s.Text())
		} else {
			// crawler not detected
			fmt.Fprintln(nonCrawlerWriter, s.Text())
		}
	}
}

func main() {
	flag.StringVar(&extraCrawlerAgentsFile, "extra-crawler-agents-file", "", "File containing additional crawler user agent patterns, one per line")
	flag.StringVar(&userAgentKeyConfig, "user-agent-key", defaultUserAgentKey, "Json key for user agent")
	flag.StringVar(&nonCrawlerOutput, "non-crawler-output", "/dev/stdout", "File to write non-crawler output to")
	flag.StringVar(&crawlerOutput, "crawler-output", "/dev/null", "File to write crawler output to")
	flag.StringVar(&errorOutput, "error-output", "/dev/null", "File to write unparsable json iput to")
	flag.Parse()

	//use default extra agents if not specified and default file exists
	if len(extraCrawlerAgentsFile) == 0 {
		if _, err := os.Stat(defaultExtraCrawlerAgentsFile); err == nil {
			extraCrawlerAgentsFile = defaultExtraCrawlerAgentsFile
		}
	}

	//load extra agents file if set
	if len(extraCrawlerAgentsFile) > 0 {
		if _, err := os.Stat(extraCrawlerAgentsFile); err == nil {
			extraAgents, err := os.Open(extraCrawlerAgentsFile)
			if err != nil {
				fmt.Println(err)
			}
			defer extraAgents.Close()

			addExtraCrawlerAgents(extraAgents)
		} else {
			fmt.Println("Error loading extra agents file", extraCrawlerAgentsFile, err)
		}
	}

	nonCrawlerWriter := getWriter(nonCrawlerOutput)
	defer nonCrawlerWriter.Close()
	crawlerWriter := getWriter(crawlerOutput)
	defer crawlerWriter.Close()
	errorWriter := getWriter(errorOutput)
	defer errorWriter.Close()

	cleanCrawlers(userAgentKeyConfig, os.Stdin, nonCrawlerWriter, crawlerWriter, errorWriter)
}
