package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Information struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	Date     string `json:"date"`
	Url      string `json:"url"`
	Location string `json:"location"` // 開催場所
	Category string `json:"category"` // イベントカテゴリ
	Source   string `json:"source"`   // 情報ソース
}

func getDoc(url string) *goquery.Document {
	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("HTTP GET エラー [%s]: %v\n", url, err)
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		// 404は多発するので詳細ログは出さない
		return nil
	}

	if res.StatusCode != 200 {
		fmt.Printf("HTTPステータスエラー [%s]: %d %s\n", url, res.StatusCode, res.Status)
		return nil
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Printf("HTML解析エラー [%s]: %v\n", url, err)
		return nil
	}

	return doc
}

// extractEventFromPage は個別ページからイベント情報を抽出する（柔軟な構造対応）
func extractEventFromPage(doc *goquery.Document, pageURL string) *Information {
	if doc == nil {
		return nil
	}

	// ページ全体のテキストから松原団地記念公園を含むか確認
	pageText := doc.Text()
	if !strings.Contains(pageText, "松原団地記念公園") {
		return nil
	}

	var title string
	var content string
	var date string
	var location string

	// タイトルを複数の方法で取得
	title = strings.TrimSpace(doc.Find("h1 span").First().Text())
	if title == "" {
		title = strings.TrimSpace(doc.Find("h1").First().Text())
	}
	if title == "" {
		title = strings.TrimSpace(doc.Find("title").First().Text())
	}

	// h2見出しから情報を収集（より柔軟に）
	doc.Find("h2").Each(func(i int, h2 *goquery.Selection) {
		h2Text := strings.TrimSpace(h2.Text())

		// 次の要素を取得（p, ul, divなど）
		nextElem := h2.Next()
		var elemText string

		// 次の要素のテキストを取得
		if nextElem.Length() > 0 {
			elemText = strings.TrimSpace(nextElem.Text())
		}

		// h2の後ろにあるすべてのp要素を取得
		if elemText == "" {
			h2.NextAll().Each(func(j int, s *goquery.Selection) {
				if s.Is("p") && elemText == "" {
					elemText = strings.TrimSpace(s.Text())
				}
			})
		}

		// 内容の抽出
		if strings.Contains(h2Text, "内容") && elemText != "" {
			content = elemText
		}

		// 日程の抽出
		if (strings.Contains(h2Text, "日程") || strings.Contains(h2Text, "日時") || strings.Contains(h2Text, "とき")) && elemText != "" {
			date = elemText
		}

		// 場所の抽出
		if (strings.Contains(h2Text, "場所") || strings.Contains(h2Text, "会場") || strings.Contains(h2Text, "ところ")) && elemText != "" {
			location = elemText
		}
	})

	// テーブルからも情報を取得
	doc.Find("table tr").Each(func(i int, tr *goquery.Selection) {
		th := strings.TrimSpace(tr.Find("th").Text())
		td := strings.TrimSpace(tr.Find("td").Text())

		if td != "" {
			if strings.Contains(th, "日程") || strings.Contains(th, "日時") || strings.Contains(th, "とき") {
				date = td
			}
			if strings.Contains(th, "内容") {
				content = td
			}
			if strings.Contains(th, "場所") || strings.Contains(th, "会場") || strings.Contains(th, "ところ") {
				location = td
			}
		}
	})

	// 場所が明示されていなければ松原団地記念公園を設定
	if location == "" || !strings.Contains(location, "松原団地記念公園") {
		location = "松原団地記念公園"
	}

	// タイトルと日付があれば有効なイベントとみなす
	if title != "" && date != "" {
		return &Information{
			Title:    title,
			Content:  content,
			Date:     date,
			Url:      pageURL,
			Location: location,
			Category: "公園イベント",
			Source:   "広報そうか",
		}
	}

	return nil
}

// grep は広報そうかページからイベント情報を収集する（改善版）
func grep(url string, infos []Information) []Information {
	doc := getDoc(url)
	if doc == nil {
		fmt.Println("広報そうかページの取得に失敗しました:", url)
		return infos
	}

	// 重複チェック用のマップ
	seenURLs := make(map[string]bool)
	for _, info := range infos {
		seenURLs[info.Url] = true
	}

	// 月号のリストを取得（例: 広報そうか令和7年3月号など）
	monthIssues := make(map[string]bool)
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		linkText := strings.TrimSpace(s.Text())

		// 広報そうかの月号リンクのみを対象
		// リンクテキストに「広報そうか」が含まれ、かつhrefがディレクトリ形式（./24040199/など）
		if exists && strings.Contains(linkText, "広報そうか") && strings.Contains(href, "./") && !strings.Contains(href, "cont") && !strings.Contains(href, "/li/") {
			// 相対URLを絶対URLに変換
			var fullURL string
			if strings.HasPrefix(href, "./") {
				// ./24040199/ のような形式
				fullURL = url + href[2:]
			} else {
				return
			}

			// 末尾にスラッシュがない場合は追加
			if !strings.HasSuffix(fullURL, "/") {
				fullURL += "/"
			}

			// PDFは除外、kohosoka/r以下のみ対象
			if !strings.HasSuffix(fullURL, ".pdf") && strings.Contains(fullURL, "/kohosoka/r") {
				monthIssues[fullURL] = true
			}
		}
	})

	fmt.Printf("  月号ページ数: %d\n", len(monthIssues))

	// 各月号のページをクロール
	for issueURL := range monthIssues {
		issueDoc := getDoc(issueURL)
		if issueDoc == nil {
			continue
		}

		// 月号ページ内のすべてのリンクをチェック
		issueDoc.Find("a").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if !exists {
				return
			}

			// 広報そうか内の記事ページのみを対象（./で始まり.htmlを含む）
			if !strings.HasPrefix(href, "./") || !strings.Contains(href, ".html") {
				return
			}

			// 相対URLを絶対URLに変換
			articleURL := issueURL + href[2:]

			// 既に処理済みのURLはスキップ
			if seenURLs[articleURL] {
				return
			}

			// 記事ページを取得して解析
			articleDoc := getDoc(articleURL)
			eventInfo := extractEventFromPage(articleDoc, articleURL)
			if eventInfo != nil {
				seenURLs[articleURL] = true
				fmt.Printf("  イベント取得: %s\n", eventInfo.Title)
				infos = append(infos, *eventInfo)
			}

			time.Sleep(200 * time.Millisecond)
		})

		time.Sleep(300 * time.Millisecond)
	}

	return infos
}

// grepEventList は草加市のイベント一覧ページからイベント情報を収集する
func grepEventList(infos []Information) []Information {
	url := "https://www.city.soka.saitama.jp/li/event/index.html"
	doc := getDoc(url)
	if doc == nil {
		fmt.Println("イベント一覧ページの取得に失敗しました")
		return infos
	}

	// 重複チェック用のマップ
	seenURLs := make(map[string]bool)
	for _, info := range infos {
		seenURLs[info.Url] = true
	}

	// イベントリストから情報を抽出
	// 実際のHTML構造に合わせてセレクタを調整
	selection := doc.Find("ul.link_list > li, div.content_list > ul > li, ul.newlist > li")

	eventCount := 0
	selection.Each(func(index int, s *goquery.Selection) {
		// リンク要素を探す
		link := s.Find("a").First()
		if link.Length() == 0 {
			return
		}

		title := strings.TrimSpace(link.Text())
		eventURL, exists := link.Attr("href")

		if !exists || title == "" {
			return
		}

		// 相対URLを絶対URLに変換
		if strings.HasPrefix(eventURL, "/") {
			eventURL = "https://www.city.soka.saitama.jp" + eventURL
		} else if !strings.HasPrefix(eventURL, "http") {
			return // 不正なURL
		}

		// 既に処理済みのURLはスキップ
		if seenURLs[eventURL] {
			return
		}

		// 個別ページから詳細情報を取得
		detailDoc := getDoc(eventURL)
		if detailDoc == nil {
			return
		}

		var content string
		var date string
		var location string

		// 本文から情報を抽出（複数のセレクタを試す）
		detailDoc.Find("div.detail_free, div.main_contents, article").Each(func(i int, detail *goquery.Selection) {
			text := strings.TrimSpace(detail.Text())
			if len(text) > 0 && len(text) < 1000 { // 適度な長さの説明文
				if content == "" || (len(text) > 50 && len(text) < len(content)) {
					content = text
				}
			}
		})

		// 日程情報を探す（テーブル、定義リスト、見出しなど）
		detailDoc.Find("table tr, dl, div").Each(func(i int, elem *goquery.Selection) {
			// テーブルの場合
			th := strings.TrimSpace(elem.Find("th").Text())
			td := strings.TrimSpace(elem.Find("td").Text())

			// 定義リストの場合
			dt := strings.TrimSpace(elem.Find("dt").Text())
			dd := strings.TrimSpace(elem.Find("dd").Text())

			text := th
			value := td
			if dt != "" {
				text = dt
				value = dd
			}

			if strings.Contains(text, "日時") || strings.Contains(text, "開催日") || strings.Contains(text, "日程") {
				if value != "" {
					date = value
				}
			}
			if strings.Contains(text, "場所") || strings.Contains(text, "会場") {
				if value != "" {
					location = value
				}
			}
		})

		// 松原団地記念公園に関連するイベントのみ追加
		pageText := detailDoc.Text()
		if strings.Contains(pageText, "松原団地記念公園") {
			if date == "" {
				// ページ全体から日付らしい文字列を探す
				datePattern := regexp.MustCompile(`(\d+月\d+日[^。\n]*|令和\d+年\d+月\d+日[^。\n]*)`)
				matches := datePattern.FindStringSubmatch(pageText)
				if len(matches) > 0 {
					date = matches[0]
				}
			}

			if date != "" {
				seenURLs[eventURL] = true
				info := Information{
					Title:    title,
					Content:  content,
					Date:     date,
					Url:      eventURL,
					Location: location,
					Category: "公園イベント",
					Source:   "イベント情報",
				}
				fmt.Printf("  イベント情報から取得: %s\n", info.Title)
				infos = append(infos, info)
				eventCount++
			}
		}

		time.Sleep(300 * time.Millisecond) // サーバーへの負荷を考慮
	})

	fmt.Printf("  イベント情報ページから %d 件取得\n", eventCount)
	return infos
}

// getKohosokaUrls は広報そうかのトップページから利用可能な年度のURLを自動的に取得する
func getKohosokaUrls() []string {
	baseURL := "https://www.city.soka.saitama.jp/kohosoka/"
	doc := getDoc(baseURL)
	if doc == nil {
		fmt.Println("広報そうかトップページの取得に失敗しました")
		// フォールバック: 現在と前年のURLを返す
		return []string{
			"https://www.city.soka.saitama.jp/kohosoka/r06/",
			"https://www.city.soka.saitama.jp/kohosoka/r05/",
		}
	}

	var urls []string
	urlMap := make(map[string]bool) // 重複チェック用
	re := regexp.MustCompile(`r\d{2}`)

	// リンクから年度URLを収集
	doc.Find("a").Each(func(index int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			// r06, r05などのパターンを探す
			matches := re.FindStringSubmatch(href)
			if len(matches) > 0 {
				// 年度部分を抽出（例: r06, r05）
				year := matches[0]
				// 正しいURLを構築
				fullURL := "https://www.city.soka.saitama.jp/kohosoka/" + year + "/"

				if !urlMap[fullURL] {
					urlMap[fullURL] = true
					urls = append(urls, fullURL)
				}
			}
		}
	})

	if len(urls) == 0 {
		// URLが見つからなかった場合のフォールバック
		fmt.Println("年度URLが見つかりませんでした。デフォルトURLを使用します。")
		return []string{
			"https://www.city.soka.saitama.jp/kohosoka/r06/",
			"https://www.city.soka.saitama.jp/kohosoka/r05/",
		}
	}

	fmt.Printf("検出された年度URL: %v\n", urls)
	return urls
}

func main() {

	var infos []Information

	fmt.Println("=== 広報そうかからイベント情報を収集 ===")
	// 広報ページから年度URLを自動取得
	urls := getKohosokaUrls()
	for _, url := range urls {
		childInfo := grep(url, infos)
		infos = append(infos, childInfo...)
	}
	fmt.Printf("広報そうかから %d 件のイベントを取得\n", len(infos))

	fmt.Println("\n=== イベント情報ページから収集 ===")
	// イベント一覧ページから情報を収集
	initialCount := len(infos)
	infos = grepEventList(infos)
	fmt.Printf("イベント情報から %d 件のイベントを取得\n", len(infos)-initialCount)

	fmt.Printf("\n総計: %d 件のイベント情報を収集しました\n", len(infos))
	output, err := json.MarshalIndent(&infos, "", "\t\t")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}
	os.WriteFile("./event-grepper-app/src/park.json", output, 0644)
	fmt.Println("park.json に保存しました")
}
