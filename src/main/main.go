package main

import (
	"fmt"
	"log"
	"slices"
	"time"

	playwright "github.com/playwright-community/playwright-go"
)

// func promptForEnvironment() (string, uint16, uint8, []string, error) {
// 	var url string
// 	var sleep uint16
// 	var depth uint8
// 	var includeDomainWildcards []string

// 	fmt.Print("Enter select the crawler config (see config.json): ")

// 	// read from config.json
// 	var configFile string
// 	if _, err := fmt.Scanln(&configFile); err != nil {
// 		return "", 0, 0, nil, fmt.Errorf("could not read config file: %w", err)
// 	}

// 	// open config.json
// 	file, err := os.Open(configFile)
// 	if err != nil {
// 		return "", 0, 0, nil, fmt.Errorf("could not open config file: %w", err)
// 	}
// 	defer file.Close()

// 	// decode config.json as an array of configurations
// 	var configs []struct {
// 		URL                    string   `json:"url"`
// 		Sleep                  uint16   `json:"sleep"`
// 		Depth                  uint8    `json:"depth"`
// 		IncludeDomainWildcards []string `json:"include_domain_wildcards"`
// 	}
// 	if err := json.NewDecoder(file).Decode(&configs); err != nil {
// 		return "", 0, 0, nil, fmt.Errorf("could not decode config file: %w", err)
// 	}

// 	// List available configurations with index which can be selected by input
// 	for i, config := range configs {
// 		fmt.Printf("%d: URL: %s, Sleep: %d seconds, Depth: %d\n", i+1, config.URL, config.Sleep, config.Depth)
// 	}
// 	fmt.Print("Select a configuration by number: ")
// 	var choice int
// 	if _, err := fmt.Scanln(&choice); err != nil {
// 		return "", 0, 0, nil, fmt.Errorf("could not read choice: %w", err)
// 	}

// 	if choice < 1 || choice > len(configs) {
// 		return "", 0, 0, nil, fmt.Errorf("invalid choice: %d", choice)
// 	}

// 	// Get the selected configuration
// 	selectedConfig := configs[choice-1]
// 	url = selectedConfig.URL
// 	sleep = selectedConfig.Sleep
// 	depth = selectedConfig.Depth
// 	includeDomainWildcards = selectedConfig.IncludeDomainWildcards

// 	return url, sleep, depth, includeDomainWildcards, nil
// }

func promptForInputs() (string, uint16, uint8, []string, error) {
	var url string
	var sleep uint16
	var depth uint8
	var includeDomainWildcards []string

	fmt.Print("Enter URL to be scanned: ")
	if _, err := fmt.Scanln(&url); err != nil {
		return "", 0, 0, nil, fmt.Errorf("could not read URL: %w", err)
	}

	fmt.Print("Enter sleep duration between navigation in seconds (default 15): ")
	if _, err := fmt.Scanln(&sleep); err != nil {
		sleep = 15 // default sleep duration
	}

	fmt.Print("Enter page depth (default 1): ")
	if _, err := fmt.Scanln(&depth); err != nil {
		depth = 1 // default depth
	}

	return url, sleep, depth, includeDomainWildcards, nil
}

func addToScannedUrlsList(scannedUrlsList *[]string, url string) {
	if url != "" && !slices.Contains(*scannedUrlsList, url) {
		*scannedUrlsList = append(*scannedUrlsList, url)
	}
}

func addNetworkEventListeners(
	page *playwright.Page,
	networkRequestsList *[]string,
	networkResponsesList *[]string,
	networkResponseStatusesList *map[string]uint16,
) {
	(*page).On("request", func(request playwright.Request) {
		// if request.Method() == "GET" || request.Method() == "POST" {
		if request.URL() != "" && !slices.Contains(*networkRequestsList, request.URL()) {
			*networkRequestsList = append(*networkRequestsList, request.URL())
		}
		// }
	})
	(*page).On("response", func(response playwright.Response) {
		if response.URL() != "" && !slices.Contains(*networkResponsesList, response.URL()) {
			*networkResponsesList = append(*networkResponsesList, response.URL())
			(*networkResponseStatusesList)[response.URL()] = uint16(response.Status())
		}
	})
}

func launchBrowserAndPage() (
	*playwright.Playwright,
	*playwright.Browser,
	*playwright.Page,
	error,
) {
	if err := playwright.Install(); err != nil {
		log.Fatalf("could not install playwright: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return pw, nil, nil, fmt.Errorf("could not start playwright: %w", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return pw, nil, nil, fmt.Errorf("could not launch browser: %w", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		return pw, nil, nil, fmt.Errorf("could not create page: %w", err)
	}

	return pw, &browser, &page, nil
}

func onExit(
	pw *playwright.Playwright,
	browser *playwright.Browser,
	page *playwright.Page,
) {
	fmt.Println("Finished Does it 200 ?")
	fmt.Println("Exiting...")

	if page != nil {
		if err := (*page).Close(); err != nil {
			log.Printf("could not close page: %v", err)
		}
	}
	if browser != nil {
		if err := (*browser).Close(); err != nil {
			log.Printf("could not close browser: %v", err)
		}
	}
	if pw != nil {
		if err := (*pw).Stop(); err != nil {
			log.Printf("could not stop Playwright: %v", err)
		}
	}
}

func launchUrlInPage(
	url string,
	page *playwright.Page,
) error {
	if _, err := (*page).Goto(url); err != nil {
		return fmt.Errorf("could not goto %s: %w", url, err)
	}

	// wait for page to load
	if err := (*page).WaitForLoadState(); err != nil {
		return fmt.Errorf("could not wait for load state for %s: %w", url, err)
	}

	fmt.Printf("Launched URL: %s\n", url)

	return nil
}

func startTree(
	currentUrls []string,
	currentDepth uint8,
	sleepAmount *uint16,
	maxDepth *uint8,
	pw *playwright.Playwright,
	page *playwright.Page,
	scannedUrlsList *[]string,
	networkRequestsList *[]string,
	networkResponsesList *[]string,
	networkResponseStatusesList *map[string]uint16,
	includeDomainWildcards *[]string,
) error {
	fmt.Println("Starting tree...")

	for _, currentUrl := range currentUrls {
		if currentUrl == "" {
			continue
		}

		if (slices.Contains(*scannedUrlsList, currentUrl)) ||
			(slices.Contains(*networkRequestsList, currentUrl)) {
			fmt.Printf("Skipping already scanned URL: %s\n", currentUrl)
			continue
		}

		*scannedUrlsList = append(*scannedUrlsList, currentUrl)

		if err := launchUrlInPage(currentUrl, page); err != nil {
			log.Printf("Error launching URL or initiated download response %s: %v", currentUrl, err)
			// If this fails, it is probably a download response instead of a page load.
			continue
		}

		if *sleepAmount > 0 {
			fmt.Printf("Sleeping for %d seconds...\n", *sleepAmount)
			time.Sleep(time.Duration(*sleepAmount) * time.Second)
		}

		var childLinks []string = make([]string, 0)

		// find all hyperlinks and add href to childLinks
		links, err := (*page).Locator("a").All()
		if err != nil {
			log.Printf("could not get links: %v", err)
			continue
		}
		for _, link := range links {
			href, err := link.GetAttribute("href")
			if err != nil {
				log.Printf("could not get href attribute: %v", err)
				continue
			}

			if len(href) == 0 {
				log.Printf("href is empty for link: %v", link)
				continue
			}
			if href[0] == '/' {
				// pageUrl := (*page).URL()
				// baseUrl = pageUrl
				(*page).WaitForURL("**/*")
				baseUrl, err := (*page).Evaluate("() => window.location.origin")
				if err != nil {
					log.Printf("could not get base URL: %v", err)
					continue
				}
				if baseUrl == nil {
					log.Printf("base URL is nil")
					continue
				}
				baseUrlStr, ok := baseUrl.(string)
				if !ok {
					log.Printf("base URL is not a string: %v", baseUrl)
					continue
				}
				href = baseUrlStr + href
			}

			// Disallowed hrefs
			if href == "" || href[0:4] == "tel:" || href[0:7] == "mailto:" || href[0:11] == "javascript:" || href[0:1] == "#" {
				continue
			}

			if href != "" && !slices.Contains(childLinks, href) && !slices.Contains(*scannedUrlsList, href) {
				childLinks = append(childLinks, href)
			}
		}

		fmt.Printf("Found %d child links for %s\n", len(childLinks), currentUrl)
		fmt.Printf("Child links: %v\n", childLinks)

		fmt.Printf("Current depth is %d\n", currentDepth)

		if currentDepth <= *maxDepth {
			// TODO: Find links and iterate next level of tree using startTree
			if finish := startTree(
				// TODO: Might need to make this a pointer ? IDK
				childLinks,
				currentDepth+1,
				sleepAmount,
				maxDepth,
				pw,
				page,
				scannedUrlsList,
				networkRequestsList,
				networkResponsesList,
				networkResponseStatusesList,
				includeDomainWildcards,
			); finish != nil {
				return finish
			}
		}
	}

	return nil
}

func filterNetworkResponseStatusesListForNon200(
	networkResponseStatusesList map[string]uint16,
) map[string]uint16 {
	filteredList := make(map[string]uint16)
	for url, status := range networkResponseStatusesList {
		if status != 200 {
			filteredList[url] = status
		}
	}
	return filteredList
}

func main() {
	fmt.Println("Starting Does it 200 ?")

	pw, browser, page, err := launchBrowserAndPage()
	if err != nil {
		log.Fatalf("could not launch browser and page: %v", err)
	}

	defer onExit(pw, browser, page)

	var scannedUrlsList []string = make([]string, 0)
	var networkRequestsList []string = make([]string, 0)
	var networkResponsesList []string = make([]string, 0)
	var networkResponseStatusesList map[string]uint16 = make(map[string]uint16)

	addNetworkEventListeners(page, &networkRequestsList, &networkResponsesList, &networkResponseStatusesList)

	// rootUrl, sleepAmount, maxDepth, includeDomainWildcards, err := promptForEnvironment()
	// TODO: Warning!!! Before using this with more than 1 / 2 max depth, we need to implement safeguard includeDomainWildcards so that external domains are not crawled
	rootUrl, sleepAmount, maxDepth, includeDomainWildcards, err := promptForInputs()
	if err != nil {
		log.Fatalf("could not prompt for inputs: %v", err)
	}

	fmt.Printf("URL: %s, Sleep: %d seconds, Depth: %d\n", rootUrl, sleepAmount, maxDepth)

	var currentDepth uint8 = 0
	var currentUrl string = rootUrl

	log.Printf("Tree started")

	if finish := startTree(
		[]string{currentUrl},
		currentDepth,
		&sleepAmount,
		&maxDepth,
		pw,
		page,
		&scannedUrlsList,
		&networkRequestsList,
		&networkResponsesList,
		&networkResponseStatusesList,
		&includeDomainWildcards,
	); finish != nil {
		log.Printf("%v", finish.Error())
	}

	log.Printf("Network Requests: %v", networkRequestsList)
	// log.Printf("Network Responses: %v", networkResponsesList)
	// log.Printf("Network Response Statuses: %v", networkResponseStatusesList)

	non200Responses := filterNetworkResponseStatusesListForNon200(networkResponseStatusesList)
	if len(non200Responses) > 0 {
		fmt.Println("Non-200 responses found:")
		for url, status := range non200Responses {
			fmt.Printf("URL: %s, Status: %d\n",
				url, status)
		}
	} else {
		fmt.Println("No non-200 responses found.")
	}
}
