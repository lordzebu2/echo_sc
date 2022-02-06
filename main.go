package main

import (
	"goquery/scrapper"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

const file_Name string = "jobs.csv"

func handleHome(c echo.Context) error {
	//return c.String(http.StatusOK, "Hello World!")
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	defer os.Remove(file_Name)
	//서버에있는 jobs.csv삭제함 사용자가 파일을 다운로드 하고난 뒤에(defer)
	//요청이 다른데 같은 파일을 저장하면 좋지않기때문..

	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	//scrapper에 만들어놨던 cleanString함수를 가져와서 씀.. 첫글자 대문자..
	//ToLower는 소문자변환.. ToUpper는 대문자로 변환임

	scrapper.Scrape(term)
	return c.Attachment(file_Name, file_Name)
	//Attachment는 첨부파일을 리턴하는 기능임 jobs.csv파일을 리턴하도록 하는것 (사용자가 파일을 다운받게함)
}

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start(":1323"))
}
