package browser

import (
	"log"

	playwright "github.com/playwright-community/playwright-go"
)

// This will be a singleton class containing access to Playwright and the browser context.

type Browser struct {
	playwright *playwright.Playwright
	browser    *playwright.Browser
	page       *playwright.Page
}

func (b *Browser) Close() {
	if b.page != nil {
		if err := (*b.page).Close(); err != nil {
			log.Printf("could not close page: %v", err)
		}
	}

	if b.browser != nil {
		if err := (*b.browser).Close(); err != nil {
			log.Printf("could not close browser: %v", err)
		}
	}

	if b.playwright != nil {
		if err := (*b.playwright).Stop(); err != nil {
			log.Printf("could not stop Playwright: %v", err)
		}
	}
}

func (b *Browser) LaunchUrlInPage(
	url string,
) (*playwright.Page, error) {
	if _, err := (*b.page).Goto(url); err != nil {
		log.Printf("could not navigate to URL %s: %v", url, err)
		return nil, err
	}

	// wait for page to load
	if err := (*b.page).WaitForLoadState(); err != nil {
		log.Printf("could not wait for page to load %s: %v", url, err)
		return nil, err
	}

	log.Printf("Launched URL: %s\n", url)
	return b.page, nil
}

func (b *Browser) GetPage() *playwright.Page {
	return b.page
}

func (b *Browser) GetPlaywright() *playwright.Playwright {
	return b.playwright
}

// Warning: Need to defer Close from the main function after creating an instance of browser
func NewBrowser() (*Browser, error) {
	err := playwright.Install()
	if err != nil {
		log.Fatalf("could not install Playwright: %v", err)
		return nil, err
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start Playwright: %v", err)
		return nil, err
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
		return nil, err
	}

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create new page: %v", err)
		return nil, err
	}

	instance := &Browser{
		playwright: pw,
		browser:    &browser,
		page:       &page,
	}

	return instance, nil
}
