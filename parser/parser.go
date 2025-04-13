package parser

import (
	"fmt"
	"net/url"
	"strings"

	"database"

	"github.com/gocolly/colly"
	"gorm.io/gorm"
)

func ParseKrisha(urlStr string, db *gorm.DB, userID int64) error {
	fmt.Println("[ParseKrisha] Starting parsing process with pagination")

	allFlats, totalCount, err := parsePages(urlStr, userID)
	if err != nil {
		return err
	}

	existingFlats, err := getExistingFlats(db, userID)
	if err != nil {
		return err
	}

	flatsToCreate, flatsToDelete := processFlats(allFlats, existingFlats)

	if err := database.DeleteOldFlats(db, flatsToDelete); err != nil {
		return err
	}

	if err := saveNewFlats(db, flatsToCreate); err != nil {
		return err
	}

	fmt.Printf("[ParseKrisha] Parsing completed, found %d listings across all pages\n", totalCount)
	return nil
}

func parsePages(urlStr string, userID int64) ([]database.Flat, int, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, 0, fmt.Errorf("[parsePages] Error parsing URL: %v", err)
	}

	var allFlats []database.Flat
	totalCount := 0

	for page := 1; page <= 5; page++ {
		q := parsedURL.Query()
		q.Set("page", fmt.Sprintf("%d", page))
		parsedURL.RawQuery = q.Encode()
		pageURL := parsedURL.String()

		fmt.Printf("[parsePages] Processing page %d: %s\n", page, pageURL)

		pageFlats, pageCount, err := parsePage(pageURL, page)
		if err != nil {
			return nil, 0, fmt.Errorf("[parsePages] Error parsing page %d: %v", page, err)
		}

		for i := range pageFlats {
			pageFlats[i].UserID = uint(userID)
		}

		allFlats = append(allFlats, pageFlats...)
		totalCount += pageCount

		if pageCount == 0 {
			fmt.Printf("[parsePages] No results on page %d, stopping pagination\n", page)
			break
		}
	}

	return allFlats, totalCount, nil
}

func getExistingFlats(db *gorm.DB, userID int64) ([]database.Flat, error) {
	existingFlats, err := database.GetFlatsByUser(db, userID)
	if err != nil {
		return nil, err
	}
	return existingFlats, nil
}

func processFlats(allFlats []database.Flat, existingFlats []database.Flat) ([]database.Flat, []uint) {
	existingMap := make(map[string]database.Flat)
	for _, flat := range existingFlats {
		existingMap[flat.Link] = flat
	}

	newFlatsMap := make(map[string]database.Flat)
	for _, flat := range allFlats {
		newFlatsMap[flat.Link] = flat
	}

	var flatsToCreate []database.Flat
	for link, flat := range newFlatsMap {
		if _, exists := existingMap[link]; !exists {
			flatsToCreate = append(flatsToCreate, flat)
		}
	}

	var flatsToDelete []uint
	for link, flat := range existingMap {
		if _, exists := newFlatsMap[link]; !exists {
			flatsToDelete = append(flatsToDelete, flat.ID)
		}
	}

	return flatsToCreate, flatsToDelete
}

// Сохраняет новые flats
func saveNewFlats(db *gorm.DB, flatsToCreate []database.Flat) error {
	if len(flatsToCreate) == 0 {
		fmt.Println("[saveNewFlats] No new flats to save")
		return nil
	}
	if err := database.SaveFlats(db, flatsToCreate); err != nil {
		return fmt.Errorf("[saveNewFlats] Error saving new flats: %v", err)
	}
	fmt.Printf("[saveNewFlats] Saved %d new flats\n", len(flatsToCreate))
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
