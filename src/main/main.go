package main

import (
	"fmt"
	"log"
	"slices"
	"time"

	"doesit200/src/browser"

	"github.com/playwright-community/playwright-go"

	"doesit200/src/config"
)

func addToScannedUrlsList(scannedUrlsList *[]string, url string) {
	if url != "" && !slices.Contains(*scannedUrlsList, url) {
		*scannedUrlsList = append(*scannedUrlsList, url)
	}
}

func addNetworkEventListeners(
	browserInstance *browser.Browser,
	networkRequestsList *[]string,
	networkResponsesList *[]string,
	networkResponseStatusesList *map[string]uint16,
) {
	(*browserInstance.GetPage()).On("request", func(request playwright.Request) {
		// if request.Method() == "GET" || request.Method() == "POST" {
		if request.URL() != "" && !slices.Contains(*networkRequestsList, request.URL()) {
			*networkRequestsList = append(*networkRequestsList, request.URL())
		}
		// }
	})
	(*browserInstance.GetPage()).On("response", func(response playwright.Response) {
		if response.URL() != "" && !slices.Contains(*networkResponsesList, response.URL()) {
			*networkResponsesList = append(*networkResponsesList, response.URL())
			(*networkResponseStatusesList)[response.URL()] = uint16(response.Status())
		}
	})
}

func startTree(
	currentUrls []string,
	currentDepth uint8,
	browserInstance *browser.Browser,
	scannedUrlsList *[]string,
	networkRequestsList *[]string,
	networkResponsesList *[]string,
	networkResponseStatusesList *map[string]uint16,
	config *config.Config,
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

		if _, err := browserInstance.LaunchUrlInPage(currentUrl); err != nil {
			log.Printf("Error launching URL or initiated download response %s: %v", currentUrl, err)
			// If this fails, it is probably a download response instead of a page load.
			continue
		}

		if config.SleepAmount > 0 {
			fmt.Printf("Sleeping for %d seconds...\n", config.SleepAmount)
			time.Sleep(time.Duration(config.SleepAmount) * time.Second)
		}

		var childLinks []string = make([]string, 0)

		// find all hyperlinks and add href to childLinks
		links, err := (*browserInstance.GetPage()).Locator("a").All()
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
				(*browserInstance.GetPage()).WaitForURL("**/*")
				baseUrl, err := (*browserInstance.GetPage()).Evaluate("() => window.location.origin")
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

		if currentDepth <= *&config.MaxDepth {
			// TODO: Find links and iterate next level of tree using startTree
			if finish := startTree(
				// TODO: Might need to make this a pointer ? IDK
				childLinks,
				currentDepth+1,
				browserInstance,
				scannedUrlsList,
				networkRequestsList,
				networkResponsesList,
				networkResponseStatusesList,
				config,
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

	browserInstance, err := browser.NewBrowser()
	if err != nil {
		log.Fatalf("could not create browser: %v", err)
	}
	defer browserInstance.Close()

	var scannedUrlsList []string = make([]string, 0)
	var networkRequestsList []string = make([]string, 0)
	var networkResponsesList []string = make([]string, 0)
	var networkResponseStatusesList map[string]uint16 = make(map[string]uint16)

	addNetworkEventListeners(browserInstance, &networkRequestsList, &networkResponsesList, &networkResponseStatusesList)

	// TODO: Warning!!! Before using this with more than 1 / 2 max depth, we need to implement safeguard includeDomainWildcards so that external domains are not crawled
	config, err := config.GetConfigFromInputs()
	if err != nil {
		log.Fatalf("could not get config from inputs: %v", err)
	}

	fmt.Printf("URL: %s, Sleep: %d seconds, Depth: %d\n", config.URL, config.SleepAmount, config.MaxDepth)

	var currentDepth uint8 = 0
	var currentUrl string = config.URL

	log.Printf("Tree started")

	if finish := startTree(
		[]string{currentUrl},
		currentDepth,
		browserInstance,
		&scannedUrlsList,
		&networkRequestsList,
		&networkResponsesList,
		&networkResponseStatusesList,
		config,
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
