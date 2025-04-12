package parser

import (
	"fmt"
	"net/url"
	"strings"

	"database"

	"github.com/gocolly/colly"
	"gorm.io/gorm"
)

func ParseKrisha(urlStr string, db *gorm.DB) error {
	fmt.Println("[ParseKrisha] Starting parsing process with pagination")

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("[ParseKrisha] Error parsing URL: %v", err)
	}

	var allFlats []database.Flat
	totalCount := 0

	// Parse the first 5 pages
	for page := 1; page <= 5; page++ {
		q := parsedURL.Query()
		q.Set("page", fmt.Sprintf("%d", page))
		parsedURL.RawQuery = q.Encode()
		pageURL := parsedURL.String()

		fmt.Printf("[ParseKrisha] Processing page %d: %s\n", page, pageURL)

		pageFlats, pageCount, err := parsePage(pageURL, page)
		if err != nil {
			return fmt.Errorf("[ParseKrisha] Error parsing page %d: %v", page, err)
		}

		allFlats = append(allFlats, pageFlats...)
		totalCount += pageCount

		if pageCount == 0 {
			fmt.Printf("[ParseKrisha] No results on page %d, stopping pagination\n", page)
			break
		}
	}

	if err := database.SaveFlats(db, allFlats); err != nil {
		return fmt.Errorf("[ParseKrisha] Error saving flats to database: %v", err)
	}

	fmt.Printf("[ParseKrisha] Parsing completed, found %d listings across all pages\n", totalCount)
	return nil
}

func parsePage(pageURL string, pageNum int) ([]database.Flat, int, error) {
	c := initCollector()

	var flats []database.Flat
	count := 0

	setupCollectorHandlers(c, &count, &flats, pageNum)

	err := c.Visit(pageURL)
	if err != nil {
		return nil, 0, fmt.Errorf("[ParseKrisha] Error visiting url: %v", err)
	}

	fmt.Printf("[ParseKrisha] Page %d parsing completed, found %d listings\n", pageNum, count)
	return flats, count, nil
}

func initCollector() *colly.Collector {
	return colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"),
	)
}

func setupCollectorHandlers(c *colly.Collector, count *int, flats *[]database.Flat, pageNum int) {
	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("[ParseKrisha] Page %d: Visiting: %s\n", pageNum, r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("[ParseKrisha] Page %d: Request error: %v, status code: %d\n", pageNum, err, r.StatusCode)
	})

	c.OnHTML("div.main-col", func(e *colly.HTMLElement) {
		fmt.Printf("[ParseKrisha] Page %d: Found main column\n", pageNum)

		e.ForEach("section.a-list-with-favs div.a-card", func(_ int, card *colly.HTMLElement) {
			*count++
			processPropertyCard(card, flats, *count, pageNum)
		})
	})
}

func processPropertyCard(card *colly.HTMLElement, flats *[]database.Flat, count int, pageNum int) {
	flat := extractPropertyData(card)
	flat.Page = pageNum

	*flats = append(*flats, flat)

	fmt.Printf("[ParseKrisha] Page %d: Card %d: %s - %s - %s\n",
		pageNum, count, flat.Title, flat.Price, flat.Link)
}

func extractPropertyData(card *colly.HTMLElement) database.Flat {
	flat := database.Flat{}

	// Extract basic data
	flat.Title = card.ChildText("a.a-card__title")
	flat.Price = card.ChildText("div.a-card__price")
	flat.Location = card.ChildText("div.a-card__subtitle")
	flat.Description = card.ChildText("div.a-card__text-preview")

	link, exists := card.DOM.Find("a.a-card__title").Attr("href")
	if exists {
		flat.Link = makeAbsoluteURL(link, card.Request.URL.String())
	}

	imageURL, _ := card.DOM.Find("picture img").Attr("src")
	flat.ImageURL = imageURL

	flat.Date = card.ChildText("div.a-card__publishing-date")

	extractAdditionalParams(card, &flat)

	return flat
}

func makeAbsoluteURL(link, baseURLStr string) string {
	if strings.HasPrefix(link, "http") {
		return link
	}

	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		return link
	}

	linkURL, err := url.Parse(link)
	if err != nil {
		return link
	}

	return baseURL.ResolveReference(linkURL).String()
}

func extractAdditionalParams(card *colly.HTMLElement, flat *database.Flat) {
	card.ForEach("div.a-card__parameters span", func(_ int, param *colly.HTMLElement) {
		paramText := param.Text
		if strings.Contains(paramText, "м²") {
			flat.Area = strings.TrimSpace(paramText)
		} else if strings.Contains(paramText, "этаж") {
			flat.Floor = strings.TrimSpace(paramText)
		} else if strings.Contains(paramText, "комн") {
			flat.Rooms = strings.TrimSpace(paramText)
		}
	})
}
