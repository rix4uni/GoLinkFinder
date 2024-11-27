package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/pflag"
	"github.com/tomnomnom/gahttp"
	"github.com/rix4uni/GoLinkFinder/banner"
)

const regexStr = `(?:"|')(((?:[a-zA-Z]{1,10}://|//)[^"'/]{1,}\.[a-zA-Z]{2,}[^"']{0,})|((?:/|\.\./|\./)[^"'><,;| *()(%%$^/\\\[\]][^"'><,;|()]{1,})|([a-zA-Z0-9_\-/]{1,}/[a-zA-Z0-9_\-/]{1,}\.(?:[a-zA-Z]{1,4}|action)(?:[\?|#][^"|']{0,}|))|([a-zA-Z0-9_\-/]{1,}/[a-zA-Z0-9_\-/]{3,}(?:[\?|#][^"|']{0,}|))|([a-zA-Z0-9_\-]{1,}\.(?:php|asp|aspx|jsp|json|action|html|js|txt|xml)(?:[\?|#][^"|']{0,}|)))(?:"|')`

var founds []string

func unique(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func downloadJSFile(urls []string, concurrency int) {
	pipeLine := gahttp.NewPipeline()
	pipeLine.SetConcurrency(concurrency)
	for _, u := range urls {
		pipeLine.Get(u, gahttp.Wrap(parseFile, gahttp.CloseBody))
	}
	pipeLine.Done()
	pipeLine.Wait()
}

func parseFile(req *http.Request, resp *http.Response, err error) {
	if err != nil {
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	matchAndAdd(string(body))
}

func extractUrlFromJS(urls []string, baseUrl string) []string {
	urls = unique(urls)
	var cleaned []string

	for _, u := range urls {
		u = strings.ReplaceAll(u, "'", "")
		u = strings.ReplaceAll(u, "\"", "")

		if len(u) < 5 {
			continue
		}

		if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
			cleaned = append(cleaned, u)
		} else if strings.HasPrefix(u, "//") {
			cleaned = append(cleaned, "https:"+u)
		} else if strings.HasPrefix(u, "/") {
			cleaned = append(cleaned, baseUrl+u)
		}
	}
	return cleaned
}

func matchAndAdd(content string) []string {
	regExp, err := regexp.Compile(regexStr)
	if err != nil {
		log.Fatal(err)
	}

	links := regExp.FindAllString(content, -1)
	for _, link := range links {
		founds = append(founds, link)
	}
	return founds
}

func appendBaseUrl(urls []string, baseUrl string) []string {
	urls = unique(urls)
	var n []string
	for _, u := range urls {
		n = append(n, baseUrl+strings.TrimSpace(u))
	}
	return n
}

func extractJSLinksFromHTML(baseUrl string) []string {
	resp, err := http.Get(baseUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.Body == nil {
		log.Fatal("Response body is nil")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var htmlJS = matchAndAdd(doc.Find("script").Text())
	var urls = extractUrlFromJS(htmlJS, baseUrl)

	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		urls = append(urls, src)
	})

	urls = appendBaseUrl(urls, baseUrl)
	return urls
}

func prepareResult(result []string) []string {
	for i := range result {
		result[i] = strings.ReplaceAll(result[i], "\"", "")
		result[i] = strings.ReplaceAll(result[i], "'", "")
	}
	return result
}

func processDomain(domain string, output *string, onlyComplete, completeURL, verbose bool, concurrency int, userAgent string, delay int, timeout int) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	
	if verbose {
		fmt.Printf("Fetching: %s\n", domain)
	}

	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "https://" + domain
	}
	
	req, err := http.NewRequest("GET", domain, nil)
	if err != nil {
		log.Fatal(err)
	}

	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Delay processing if specified
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// Further processing...
	htmlUrls := extractJSLinksFromHTML(domain)
	downloadJSFile(htmlUrls, concurrency)
	founds = unique(founds)
	founds = prepareResult(founds)

	var filteredFounds []string
	for _, found := range founds {
		if onlyComplete {
			// Include only URLs that start with "http://" or "https://"
			if strings.HasPrefix(found, "http://") || strings.HasPrefix(found, "https://") {
				filteredFounds = append(filteredFounds, found)
			}
		} else {
			if completeURL {
				// Add domain to relative URLs
				if !strings.HasPrefix(found, "http://") && !strings.HasPrefix(found, "https://") {
					if strings.HasPrefix(found, "/") {
						filteredFounds = append(filteredFounds, domain+found)
					} else {
						filteredFounds = append(filteredFounds, domain+"/"+found)
					}
				} else {
					filteredFounds = append(filteredFounds, found)
				}
			} else {
				filteredFounds = append(filteredFounds, found)
			}
		}
	}

	for _, found := range filteredFounds {
		fmt.Println(found)
	}

	if *output != "" {
		f, err := os.OpenFile(*output, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()

		for _, found := range filteredFounds {
			if _, err := f.WriteString(found + "\n"); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func main() {
	domain := pflag.StringP("domain", "d", "", "Input a URL.")
	list := pflag.StringP("list", "l", "", "Input file containing a list of live subdomains to process.")
	output := pflag.StringP("output", "o", "", "File to write output results.")
	onlyComplete := pflag.Bool("only-complete", false, "Show only complete URLs starting with http:// or https://.")
	completeURL := pflag.Bool("complete-url", false, "Add the domain to relative URLs.")
	concurrency := pflag.IntP("concurrency", "c", 10, "Concurrency level.")
	userAgent := pflag.String("H", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36", "Set custom User-Agent.")
	delay := pflag.Int("delay", 0, "Delay between requests in milliseconds.")
	timeout := pflag.Int("timeout", 10, "HTTP timeout in seconds.")
	silent := pflag.Bool("silent", false, "silent mode.")
	versionFlag := pflag.Bool("version", false, "Print the version of the tool and exit.")
	verbose := pflag.Bool("verbose", false, "Enable verbose mode.")
	pflag.Parse()

	if *versionFlag {
		banner.PrintBanner()
		banner.PrintVersion()
		return
	}

	if !*silent {
		banner.PrintBanner()
	}

	if *domain != "" {
		processDomain(*domain, output, *onlyComplete, *completeURL, *verbose, *concurrency, *userAgent, *delay, *timeout)
	} else if *list != "" {
		file, err := os.Open(*list)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			processDomain(scanner.Text(), output, *onlyComplete, *completeURL, *verbose, *concurrency, *userAgent, *delay, *timeout)
		}
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			reader := bufio.NewReader(os.Stdin)
			for {
				line, err := reader.ReadString('\n')
				if err == io.EOF {
					break
				}
				processDomain(strings.TrimSpace(line), output, *onlyComplete, *completeURL, *verbose, *concurrency, *userAgent, *delay, *timeout)
			}
		} else {
			fmt.Println("Usage: Provide a domain (-d), list (-l), or input via stdin.")
		}
	}
}
