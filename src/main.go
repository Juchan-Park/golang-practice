package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	iconv "github.com/djimenez/iconv-go"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id       string
	title    string
	location string
	salary   string
}

var baseUrl string = "https://search.incruit.com/list/search.asp?col=job&kw=python"

func main() {
	var jobs []extractedJob
	totalPages := getPages()
	for i := 0; i < totalPages; i++ {
		extractedJob := getPage(i)
		jobs = append(jobs, extractedJob...)
	}

	fmt.Println(jobs)
	writeJobs(jobs)
	fmt.Println("Done. Extracted: ", len(jobs))
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv") //파일생성
	checkErr(err)

	w := csv.NewWriter(file) //연필생성
	defer w.Flush()

	headers := []string{"id", "title", "location"} //헤더생성
	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs { //각 정보마다 Write함수로 작성 후 Flush로 저장
		jobSlice := []string{job.id, job.title, job.location}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}

}

func getPage(page int) []extractedJob {
	var jobs []extractedJob
	pageUrl := baseUrl + "&startno=" + strconv.Itoa(page*30)
	fmt.Println("Requesting: ", pageUrl)

	res, err := http.Get(pageUrl)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".c_row").Each(func(i int, card *goquery.Selection) {
		job := extractJob(card)
		jobs = append(jobs, job)
	})

	return jobs

}

func extractJob(card *goquery.Selection) extractedJob {
	id, _ := card.Attr("jobno")
	title := card.Find(".cl_top>a").Text()
	out, _ := iconv.ConvertString(string(title), "euc-kr", "utf-8")
	cleanSpace(out)

	location := card.Find(".cl_md span").Text()
	out1, _ := iconv.ConvertString(string(location), "euc-kr", "utf-8")
	cleanSpace(out1)

	return extractedJob{
		id:       id,
		title:    out,
		location: out1,
	}

}

func cleanSpace(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func getPages() int {
	pages := 0
	res, err := http.Get(baseUrl)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".sqr_paging").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length() - 1
	})

	return pages
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
		fmt.Println(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with statusCode:", res.StatusCode)
	}
}
