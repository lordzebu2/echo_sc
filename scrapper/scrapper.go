package scrapper

/*
고루틴 으로 바꾼다..
지금은 getPage가 한번 요청한 뒤 다음 페이지를 요청하는 방식..
고루틴으로 페이지들과 job정보들을 동시에 요청하게끔 바꿔준다

1. 총 페이지 수를 가져온다.
2. 각 페이지 별로 고루틴 생성
3. getPage는 각 일자리 정보별로 고루틴 생성
 일자리 정보(extractJob) - getPage // getPage - main 채널로 연결
*/
import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	//go mod tidy <== 모듈을 추가/설치하면 유효성 검사를 해줘야함...
)

type extractedJob struct {
	id       string
	location string
	title    string
	dDay     string
	summary  string
}

// Scrape Indeed by a term
func Scrape(term string) {
	var baseURL string = "https://kr.indeed.com/jobs?q=" + term + "&l="
	var jobs []extractedJob //getPage는 jobs의 배열을 return하기 때문에 main에도 만들어줘야함 main의 jobs는 여러 배열들의 조합이다..
	mainC := make(chan []extractedJob)
	totalPages := getPages(baseURL)
	//페이지 수 가져옴

	for i := 0; i < totalPages; i++ {
		go getPage(i, baseURL, mainC)
		//각 페이지 별로 getPage함수를 호출
		//getPage는 각 페이지에 있는 일자리를 모두 반환(jobs)하는 함수

		//여러배열들 [](slice) 들을 하나로 만드는 작업
	}

	for i := 0; i < totalPages; i++ {
		extractedJob := <-mainC
		jobs = append(jobs, extractedJob...)
		//extractedJob을 getPage의 리턴된 값에서 가져오는 대신 채널로 전달된 메세지를 사용한다.
	}

	writeJobs(jobs)
	fmt.Println("Done, extracted", len(jobs))
}

func getPage(page int, url string, mainC chan<- []extractedJob) {
	//getPage 에서는 필요한 주소를 만들고 로그인을 한 후에 getPages에서 한 것처럼 정보를 가져오는 요청을 한다.
	var jobs []extractedJob
	c := make(chan extractedJob)
	pageURL := url + "&start=" + strconv.Itoa(page*10)
	//fmt.Println("Requesting", pageURL, "리퀘스팅 페이지URL TEST")

	res, err := http.Get(pageURL) //요청
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()
	//body는 io함수(input/output)이기 때문에  쓰면 닫아줘야함(getpages가 끝날때)

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".tapItem")
	//모든 카드를 찾고 Each로 각 카드의 일자리 정보들을 가져온다.
	searchCards.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
		//jobs에 job을 더해주고 jobs에 저장
		//extractJob에서 추출한 struct를 jobs에 저장후 main으로 return.
	})

	for i := 0; i < searchCards.Length(); i++ {
		//전달받을 메세지의 숫자는 카드의 갯수와 같다.
		//extractJob 함수는 각 카드마다 1번씩 실행되지 때문임
		job := <-c
		jobs = append(jobs, job)
		//메세지가 전달되기를 기다렸다가 메세지를 받으면 jobs배열에 추가한다.
	}
	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("data-jk")
	title := CleanString(card.Find(".jobTitle>span").Text())
	location := CleanString(card.Find(".companyLocation").Text())
	dDay := CleanString(card.Find(".date").Text())
	summary := CleanString(card.Find(".job-snippet").Text())

	c <- extractedJob{
		id:       id,
		title:    title,
		location: location,
		dDay:     dDay,
		summary:  summary,
	}
}

// CleanString cleans a string
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
	//띄어쓰기 공백 없에주는 string함수
	//Fields는 string주의를 분리 하는 함수임 텍스트로만 이루어진 배열(slice)을 만들어줌
	//Join은 배열을 하나의 string으로 합쳐줌
	//" " 는 너무 다닥다닥 붙어있으면 보기 않좋으니 띄움
	//결과는 불필요한 공백을 없에고 이것을 하나의 string으로 합침
}

func getPages(url string) int {
	pages := 0
	res, err := http.Get(url) //baseURL에 데이터를 요청함
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()
	//body는 io함수(input/output)이기 때문에  쓰면 닫아줘야함(getpages가 끝날때)

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func writeJobs(jobs []extractedJob) {
	//이함수는 일자리를 csv파일로 저장하는 역할
	file, err := os.Create("jobs.csv")
	//jobs.csv파일 생성
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()
	//Flush로 실행함 함수가 끝나는 시점에 파일에 데이터를 입력한다.

	headers := []string{"Link", "Title", "Location", "dDay", "Summary"}
	//파일의 맨위 속성값들
	wErr := w.Write(headers)
	//입력
	checkErr(wErr)
	//Write함수는 err를 리턴한다.

	for _, job := range jobs {
		jobSlice := []string{"https://kr.indeed.com/jobs?q=python&vjk=" + job.id, job.title, job.location, job.dDay, job.summary}
		jwErr := w.Write(jobSlice)
		//입력
		checkErr(jwErr)
	}
	//for가 끝나면 defer가 실행되고 데이터가 파일에 입력된다.
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}
