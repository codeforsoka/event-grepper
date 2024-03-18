package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Information struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Date    string `json:"date"`
	Url     string `json:"url"`
}

func getDoc(url string) *goquery.Document {
	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		fmt.Println("status code 404:", res)
		return nil
	}

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
		return nil
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	return doc
}

func main() {

	var infos []Information

	// 令和5年広報ページの内容を取得
	url := "https://www.city.soka.saitama.jp/kohosoka/r05/"
	doc := getDoc(url)
	if doc == nil {
		fmt.Println("Can not get infomation")
		return
	}
	selection := doc.Find("div#main > div.inside > div.contents_wrap > article.article > section.menu_section > div.section_wrap > ul.menu_list > li")
	selection.Each(func(index int, s *goquery.Selection) {
		selection := s.Find("a")
		attr, exists := selection.Attr("href")
		fmt.Println("test: 1")
		if exists {
			url2 := url + attr[2:]
			doc := getDoc(url2)
			fmt.Println("test: 2")
			// 広報そうか令和ページの内容を取得
			if doc != nil && strings.Contains(doc.Find("title").Text(), "広報そうか令和") {
				fmt.Println("test: 3")
				selection := doc.Find("div#main > div.inside > div.contents_wrap > article.article > div.txtbox > ul > li")
				selection.Each(func(index int, s *goquery.Selection) {
					selection := s.Find("a")
					attr, exists := selection.Attr("href")
					fmt.Println("test: 4")
					if exists {
						// 松原団地記念公園の内容が書かれたページの内容を取得
						url3 := url2 + attr[2:]
						fmt.Println("test: 5")
						doc := getDoc(url3)
						if doc != nil && strings.Contains(doc.Text(), "松原団地記念公園") {
							var title string
							var content string
							var date string
							fmt.Println("test: 6")
							title = doc.Find("div#main > div.inside > h1 > span").Text()
							// 内容と日程情報を取得
							selection_header := doc.Find("div#main > div.inside > div.contents_wrap > article.article > div.txtbox > h2")
							selection_header.Each(func(index int, s *goquery.Selection) {
								h2Text := s.Text()
								pText := s.NextFiltered("p").Text()
								if pText != "" {
									// 対応するpタグが見つかった場合のみ、結果に追加
									if strings.Contains(h2Text, "内容") {
										content = pText
									} else if strings.Contains(h2Text, "日程") {
										date = pText
									}
								}
							})

							if content != "" && date != "" {
								info := Information{
									Title:   title,
									Content: content,
									Date:    date,
									Url:     url3,
								}
								fmt.Println("info:", info)
								infos = append(infos, info)
							} else {
								fmt.Println("not content:", url3)
							}
						}
						time.Sleep(200 * time.Millisecond)
					}
				})
			}
			time.Sleep(200 * time.Millisecond)
		}
	})

	fmt.Println("infos:", infos)
	output, err := json.MarshalIndent(&infos, "", "\t\t")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}
	os.WriteFile("park.json", output, 0644)
}
