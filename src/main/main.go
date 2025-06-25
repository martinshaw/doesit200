package main

import (
	"fmt"
	"log"
	"slices"
	"time"

	playwright "github.com/playwright-community/playwright-go"
)

func promptForInputs() (string, uint16, uint8, error) {
	var url string
	var sleep uint16
	var depth uint8

	fmt.Print("Enter URL to be scanned: ")
	if _, err := fmt.Scanln(&url); err != nil {
		return "", 0, 0, fmt.Errorf("could not read URL: %w", err)
	}

	fmt.Print("Enter sleep duration between navigation in seconds (default 15): ")
	if _, err := fmt.Scanln(&sleep); err != nil {
		sleep = 15 // default sleep duration
	}

	fmt.Print("Enter page depth (default 1): ")
	if _, err := fmt.Scanln(&depth); err != nil {
		depth = 1 // default depth
	}

	return url, sleep, depth, nil
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

	playwright, err := playwright.Run()
	if err != nil {
		return playwright, nil, nil, fmt.Errorf("could not start playwright: %w", err)
	}

	browser, err := playwright.Chromium.Launch()
	if err != nil {
		return playwright, nil, nil, fmt.Errorf("could not launch browser: %w", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		return playwright, nil, nil, fmt.Errorf("could not create page: %w", err)
	}

	return playwright, &browser, &page, nil
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

	fmt.Printf("Launched URL: %s\n", url)

	title, err := (*page).Title()
	if err != nil {
		return fmt.Errorf("could not get title for %s: %w", url, err)
	}
	fmt.Printf("Title for %s: %s\n", url, title)

	return nil
}

func startTree(
	currentUrls []string,
	currentDepth *uint8,
	sleepAmount *uint16,
	maxDepth *uint8,
	page *playwright.Page,
) {
	fmt.Println("Starting tree...")

	if *currentDepth > *maxDepth {
		fmt.Printf("Reached max depth of %d, stopping...\n", *maxDepth)
		return
	}

	for _, currentUrl := range currentUrls {
		if currentUrl == "" {
			continue
		}

		if err := launchUrlInPage(currentUrl, page); err != nil {
			log.Printf("Error launching URL %s: %v", currentUrl, err)
			return
		}

		if *sleepAmount > 0 {
			fmt.Printf("Sleeping for %d seconds...\n", *sleepAmount)
			time.Sleep(time.Duration(*sleepAmount) * time.Second)
		}

		(*currentDepth)++

		// TODO: Find links and iterate next level of tree using startTree
		startTree(
			[]string{""},
			currentDepth,
			sleepAmount,
			maxDepth,
			page,
		)
	}
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

	var networkRequestsList []string = make([]string, 0)
	var networkResponsesList []string = make([]string, 0)
	var networkResponseStatusesList map[string]uint16 = make(map[string]uint16)

	addNetworkEventListeners(page, &networkRequestsList, &networkResponsesList, &networkResponseStatusesList)

	rootUrl, sleepAmount, maxDepth, err := promptForInputs()
	if err != nil {
		log.Fatalf("could not prompt for inputs: %v", err)
	}

	fmt.Printf("URL: %s, Sleep: %d seconds, Depth: %d\n", rootUrl, sleepAmount, maxDepth)

	var currentDepth uint8 = 0
	var currentUrl string = rootUrl

	log.Printf("Tree started")

	startTree(
		[]string{currentUrl},
		&currentDepth,
		&sleepAmount,
		&maxDepth,
		page,
	)

	log.Printf("Network Requests: %v", networkRequestsList)
	log.Printf("Network Responses: %v", networkResponsesList)
	log.Printf("Network Response Statuses: %v", networkResponseStatusesList)

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

// if _, err = (*page).Goto(url); err != nil {
// 	log.Fatalf("could not goto: %v", err)
// }

// entries, err := (*page).Locator(".athing").All()
// if err != nil {
// 	log.Fatalf("could not get entries: %v", err)
// }
// for i, entry := range entries {
// 	title, err := entry.Locator("td.title > span > a").TextContent()
// 	if err != nil {
// 		log.Fatalf("could not get text content: %v", err)
// 	}
// 	fmt.Printf("%d: %s\n", i+1, title)
// }
