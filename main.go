package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"net/http"
)

type Information struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Date     string   `json:"date"`
	Url      string   `json:"url"`
	Location string   `json:"location"`
	Category string   `json:"category"`
	Source   string   `json:"source"`
	Tags     []string `json:"tags,omitempty"` // サブカテゴリ・タグ
}

// 日本語の日付をYYYY-MM-DD形式に変換
func parseJapaneseDate(dateStr string) string {
	// 既にYYYY-MM-DD形式の場合はそのまま返す
	if regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`).MatchString(dateStr) {
		return dateStr
	}

	currentYear := time.Now().Year()

	// パターン1: "2025年1月15日" -> "2025-01-15"
	re1 := regexp.MustCompile(`(\d{4})年(\d{1,2})月(\d{1,2})日`)
	if matches := re1.FindStringSubmatch(dateStr); matches != nil {
		year := matches[1]
		month := matches[2]
		day := matches[3]
		monthInt, _ := strconv.Atoi(month)
		dayInt, _ := strconv.Atoi(day)
		return fmt.Sprintf("%s-%02d-%02d", year, monthInt, dayInt)
	}

	// パターン2: "1月15日" -> "2025-01-15" (現在の年を使用)
	re2 := regexp.MustCompile(`(\d{1,2})月(\d{1,2})日`)
	if matches := re2.FindStringSubmatch(dateStr); matches != nil {
		month, _ := strconv.Atoi(matches[1])
		day, _ := strconv.Atoi(matches[2])

		// 月が1-3月で現在が10-12月の場合、来年とみなす
		now := time.Now()
		year := currentYear
		if month >= 1 && month <= 3 && now.Month() >= 10 {
			year++
		}

		return fmt.Sprintf("%d-%02d-%02d", year, month, day)
	}

	// パターン3: "令和7年1月15日" -> "2025-01-15"
	re3 := regexp.MustCompile(`令和(\d+)年(\d{1,2})月(\d{1,2})日`)
	if matches := re3.FindStringSubmatch(dateStr); matches != nil {
		reiwaYear, _ := strconv.Atoi(matches[1])
		year := 2018 + reiwaYear // 令和元年 = 2019年
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		return fmt.Sprintf("%d-%02d-%02d", year, month, day)
	}

	// パターンマッチしない場合は元の文字列を返す
	return dateStr
}

func getDoc(url string) *goquery.Document {
	res, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil
	}

	return doc
}

// やしおんのイベントリストページから個別イベントURLを取得
func getYashionEventURLs() []string {
	var urls []string

	// 最初の3ページを取得（ページネーション対応）
	for page := 1; page <= 3; page++ {
		var pageURL string
		if page == 1 {
			pageURL = "https://yashion.jp/event/"
		} else {
			pageURL = fmt.Sprintf("https://yashion.jp/event/page/%d/", page)
		}

		doc := getDoc(pageURL)
		if doc == nil {
			fmt.Printf("ページ取得失敗: %s\n", pageURL)
			continue
		}

		// イベントリンクを抽出
		doc.Find("a[href*='/event/']").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists && strings.Contains(href, "/event/") && strings.Contains(href, "yashion.jp") {
				// /event/XXXXX/ 形式のURLのみを抽出
				if matched, _ := regexp.MatchString(`/event/\d+/$`, href); matched {
					urls = append(urls, href)
				}
			}
		})

		time.Sleep(500 * time.Millisecond) // レート制限対策
	}

	// 重複を除去
	seen := make(map[string]bool)
	var uniqueURLs []string
	for _, url := range urls {
		if !seen[url] {
			seen[url] = true
			uniqueURLs = append(uniqueURLs, url)
		}
	}

	return uniqueURLs
}

// やしおんの個別イベントページから情報を抽出
func extractYashionEvent(eventURL string) *Information {
	doc := getDoc(eventURL)
	if doc == nil {
		return nil
	}

	info := &Information{
		Url:    eventURL,
		Source: "やしおん",
	}

	// タイトルを取得
	title := doc.Find("h1").First().Text()
	info.Title = strings.TrimSpace(title)

	// 日付を取得（例: "2025年12月8日(月)"）
	dateText := ""
	doc.Find("div, span, p").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if regexp.MustCompile(`\d{4}年\d{1,2}月\d{1,2}日`).MatchString(text) {
			dateText = text
		}
	})

	if dateText != "" {
		// 日付部分のみを抽出
		re := regexp.MustCompile(`(\d{4}年\d{1,2}月\d{1,2}日)`)
		if matches := re.FindStringSubmatch(dateText); matches != nil {
			info.Date = parseJapaneseDate(matches[1])
		}
	}

	// カテゴリとタグを取得
	var tags []string
	doc.Find("a[href*='/event-taxonomy/']").Each(func(i int, s *goquery.Selection) {
		tag := strings.TrimSpace(s.Text())
		if tag != "" {
			tags = append(tags, tag)
		}
	})

	if len(tags) > 0 {
		info.Category = tags[0] // 最初のタグをメインカテゴリに
		if len(tags) > 1 {
			info.Tags = tags[1:] // 残りをタグに
		}
	}

	// 本文から場所を抽出（簡易版）
	contentText := doc.Find("article, .entry-content, .post-content").Text()
	info.Content = strings.TrimSpace(contentText)

	// 本文が長すぎる場合は最初の200文字に制限
	if len(info.Content) > 200 {
		info.Content = info.Content[:200] + "..."
	}

	// 場所を抽出（"場所:"や"会場:"の後の文字列）
	locationPatterns := []string{
		`場所[：:]\s*([^\n]+)`,
		`会場[：:]\s*([^\n]+)`,
		`開催場所[：:]\s*([^\n]+)`,
	}

	for _, pattern := range locationPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(contentText); matches != nil && len(matches) > 1 {
			info.Location = strings.TrimSpace(matches[1])
			break
		}
	}

	// タイトルと日付が必須
	if info.Title == "" || info.Date == "" {
		return nil
	}

	return info
}

func main() {
	var infos []Information
	seenURLs := make(map[string]bool)

	fmt.Println("=== やしおんからイベント情報を収集 ===")

	// やしおんからイベントURLリストを取得
	fmt.Println("\nイベントリストを取得中...")
	eventURLs := getYashionEventURLs()
	fmt.Printf("取得したイベントURL数: %d\n", len(eventURLs))

	// 並列処理でイベント情報を取得（5並列）
	semaphore := make(chan struct{}, 5)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, eventURL := range eventURLs {
		// 最大50件まで処理（制限を設ける）
		if i >= 50 {
			break
		}

		wg.Add(1)
		go func(url string, index int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fmt.Printf("  [%d/%d] 処理中: %s\n", index+1, len(eventURLs), url)

			info := extractYashionEvent(url)
			if info != nil {
				mu.Lock()
				if !seenURLs[info.Url] {
					seenURLs[info.Url] = true
					infos = append(infos, *info)
					fmt.Printf("    ✓ イベント取得: %s (日付: %s, カテゴリ: %s)\n", info.Title, info.Date, info.Category)
				}
				mu.Unlock()
			}

			time.Sleep(200 * time.Millisecond) // レート制限対策
		}(eventURL, i)
	}

	wg.Wait()

	fmt.Printf("\n総計: %d 件のイベント情報を収集しました\n", len(infos))

	if len(infos) == 0 {
		fmt.Println("警告: イベント情報が1件も取得できませんでした")
	}

	// カテゴリ別の集計を表示
	categoryCount := make(map[string]int)
	for _, info := range infos {
		categoryCount[info.Category]++
	}

	fmt.Println("\n=== カテゴリ別集計 ===")
	for category, count := range categoryCount {
		fmt.Printf("  %s: %d件\n", category, count)
	}

	output, err := json.MarshalIndent(&infos, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}
	os.WriteFile("./event-grepper-app/src/park.json", output, 0644)
	fmt.Println("\npark.json に保存しました")
}
