package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"

	"github.com/labstack/echo/v4"
	"github.com/mmcdole/gofeed"
)

type nwsItem struct {
	Title    string
	Href     string
	Content  string
	ImageURL string
}

func grabFeed() (*gofeed.Feed, error) {
	fp := gofeed.NewParser()
	return fp.ParseURL("https://www.vrt.be/vrtnws/nl.rss.articles.xml")
}

func serveNewsList(c echo.Context) error {
	feed, err := grabFeed()
	if err != nil {
		return c.String(http.StatusInternalServerError, "")
	}

	maxItems := 30
	if c.QueryParam("max") != "" {
		maxItems, err = strconv.Atoi(c.QueryParam("max"))
		if err != nil {
			return c.String(http.StatusBadRequest, "")
		}
	}

	tmpl := template.Must(template.ParseFiles("./static/nws/list.wml"))

	c.Response().Header().Set("Content-Type", "text/vnd.wap.wml")
	c.Response().Header().Set("Cache-Control", "no-cache, must-revalidate")

	nwsItems := []nwsItem{}
	for _, item := range feed.Items {
		if strings.HasPrefix(item.Title, "Het weer") || strings.HasPrefix(item.Title, "Het Journaal") {
			// useless news items
			continue
		}
		nwsItems = append(nwsItems, nwsItem{
			Title: trimTitle(item.Title),
			Href:  fmt.Sprintf("/nws/item?id=%s", template.URLQueryEscaper(item.GUID)),
		})
	}

	if len(nwsItems) > maxItems {
		nwsItems = nwsItems[:maxItems]
	}

	return tmpl.Execute(c.Response().Writer, struct{ Items []nwsItem }{Items: nwsItems})
}

func serveNewsItem(c echo.Context) error {
	feed, err := grabFeed()
	if err != nil {
		return c.String(http.StatusInternalServerError, "")
	}

	tmpl := template.Must(template.ParseFiles("./static/nws/item.wml"))

	id := c.QueryParam("id")
	var article *gofeed.Item

	for _, item := range feed.Items {
		if item.GUID == id {
			article = item
		}
	}

	if article == nil {
		log.Println(id)
		return c.String(http.StatusNotFound, "")
	}

	item := nwsItem{
		Title:   trimTitle(article.Title),
		Content: fmt.Sprintf("%s<br />%s", article.Title, article.Description),
	}

	if article.Image != nil {
		item.ImageURL = article.Image.URL

	} else if len(article.Enclosures) > 0 {
		for _, enclosure := range article.Enclosures {
			if strings.HasPrefix(enclosure.Type, "image") {
				item.ImageURL = enclosure.URL
				break
			}
		}
	}

	if item.ImageURL != "" {
		cachedLink := StoreLink(item.ImageURL)
		item.ImageURL = fmt.Sprintf("/png-convert.wbmp?url=%s", url.QueryEscape("cache:"+cachedLink))
	}

	c.Response().Header().Set("Content-Type", "text/vnd.wap.wml")
	err = tmpl.Execute(c.Response().Writer, struct{ Item nwsItem }{Item: item})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func trimTitle(in string) string {
	if len(in) > 40 {
		return fmt.Sprintf("%s...", in[:40])
	}

	return in
}
