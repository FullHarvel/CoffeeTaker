package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var host string
var port string
var user string
var password string
var dbname string
var accessToken string

func setEnv() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	host = os.Getenv("HOST")
	port = os.Getenv("PORT")
	user = os.Getenv("USER")
	password = os.Getenv("PASSWORD")
	dbname = os.Getenv("DBNAME")
	accessToken = os.Getenv("ACCEESSTOKEN")
}

func main() {
	setEnv()

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static/"))))

	http.HandleFunc("/", mainPage)
	http.HandleFunc("/index", index)
	http.HandleFunc("/handler", handler)
	http.HandleFunc("/getCoffeeTime", Coffeetime)
	http.HandleFunc("/getCoffeeTimeDefault", coffeetime_default)
	http.HandleFunc("/user", userTimeSelect)
	http.HandleFunc("/userConfirm", userConfirm)

	fmt.Println("start")

	http.ListenAndServe(":8080", nil)
}

func sendLineMessage(message string) {

	lineAccessToken := accessToken
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Now().In(jst).Format("2006-01-02")
	lineMessage := now + "日の" + message + "時間に予約が入りました。"
	URL := "https://notify-api.line.me/api/notify"

	u, err := url.ParseRequestURI(URL)
	if err != nil {
		log.Fatal(err)
	}

	c := &http.Client{}
	form := url.Values{}
	form.Add("message", lineMessage)
	body := strings.NewReader(form.Encode())

	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+lineAccessToken)

	_, err = c.Do(req)
	if err != nil {
		log.Fatal(err)
	}
}

func getUsersTime() ([]Record, error) {
	db, err := ConnectDB()
	defer db.Close()

	var records []Record
	var r Record
	isTodayData, starttime, finishtime := isTodayData()

	if isTodayData {
		r.Starttime = starttime
		r.Finishtime = finishtime
		records = append(records, r)
		return records, nil
	}

	//coffeetime_defaultから値を取ってくる
	rows, err := db.Query("SELECT starttime, finishtime FROM coffeetime_default order by id desc limit 1")
	if err != nil {
		return nil, fmt.Errorf("failed to query the database: %w", err)
	}
	defer rows.Close()

	for rows.Next() {

		if err := rows.Scan(&r.Starttime, &r.Finishtime); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		records = append(records, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	return records, nil
}

func changetime(starttime string, finishtime string) []string {
	allTimes := []string{
		"00:00", //0
		"00:30", //1
		"01:00", //2
		"01:30", //3
		"02:00", //4
		"02:30", //5
		"03:00", //6
		"03:30", //7
		"04:00", //8
		"04:30", //9
		"05:00", //10
		"05:30", //11
		"06:00", //12
		"06:30", //13
		"07:00", //14
		"07:30", //15
		"08:00", //16
		"08:30", //17
		"09:00", //18
		"09:30", //19
		"10:00", //20
		"10:30", //21
		"11:00", //22
		"11:30", //23
		"12:00", //24
		"12:30", //25
		"13:00", //26
		"13:30", //27
		"14:00", //28
		"14:30", //29
		"15:00", //30
		"15:30", //31
		"16:00", //32
		"16:30", //33
		"17:00", //34
		"17:30", //35
		"18:00", //36
		"18:30", //37
		"19:00", //38
		"19:30", //39
		"20:00", //40
		"20:30", //41
		"21:00", //42
		"21:30", //43
		"22:00", //44
		"22:30", //45
		"23:00", //46
		"23:30", //47
	}

	startIndex := makeTimeId(allTimes, starttime)
	finishIndex := makeTimeId(allTimes, finishtime)

	result := getTime(allTimes, startIndex, finishIndex)

	return result
}

func makeTimeId(allTime []string, time string) int {
	for i, v := range allTime {
		if v == time {
			return i
		}
	}
	return -1
}

func getTime(allTimes []string, startIndex int, finishIndex int) []string {
	var selectedTimes []string
	for i := startIndex; i <= finishIndex; i++ {
		selectedTimes = append(selectedTimes, allTimes[i])
	}
	return selectedTimes
}

func userConfirm(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	message := r.Form.Get("message")
	sendLineMessage(message)
	http.ServeFile(w, r, "user_conf.html")

}

type IsTodayData struct {
	Today      time.Time
	Starttime  string
	Finishtime string
}

func isTodayData() (bool, string, string) {
	db, err := ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Now().In(jst).Format("2006-01-02")

	rows, err := db.Query("SELECT created_at, starttime, finishtime  FROM coffeetime order by id desc limit 1")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var is_today_data IsTodayData
		if err := rows.Scan(&is_today_data.Today, &is_today_data.Starttime, &is_today_data.Finishtime); err != nil {
			log.Fatal(err)
		}

		if now == is_today_data.Today.Format("2006-01-02") {
			return true, is_today_data.Starttime, is_today_data.Finishtime
		}
	}

	return false, "", ""
}

type Data struct {
	Message string
}

func renderTemplate(w http.ResponseWriter, tmpl string, data Data) {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ConnectDB() (*sql.DB, error) {
	port, _ := strconv.Atoi(port)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func InsertTimeData(db *sql.DB, tableName string, data1 string, data2 string) error {
	// SQL文の作成
	sqlStatement := `
        INSERT INTO ` + tableName + `(starttime,finishtime)
        VALUES ($1, $2)`
	starttime := data1
	finishtime := data2

	_, err := db.Exec(sqlStatement, starttime, finishtime)
	if err != nil {
		return err
	}
	return nil
}

type Record struct {
	Today      time.Time
	Starttime  string
	Finishtime string
}

func getUsers(tableName string) ([]Record, error) {
	db, err := ConnectDB()
	defer db.Close()

	queryDocument := "SELECT created_at,starttime, finishtime FROM " + tableName + " ORDER BY id desc"
	rows, err := db.Query(queryDocument)
	if err != nil {
		return nil, fmt.Errorf("failed to query the database: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.Today, &r.Starttime, &r.Finishtime); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		records = append(records, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	return records, nil
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "main.html")
}

func index(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", Data{Message: ""})
	return
}

type Time struct {
	Starttime  string
	Finishtime string
}

func handler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	judgeDataForm := r.Form.Get("judgeDataForm")

	starttime := r.Form.Get("starttime")
	finishtime := r.Form.Get("finishtime")
	tableName := convertTableName(judgeDataForm)

	date1, _ := time.Parse("15:04", starttime)
	date2, _ := time.Parse("15:04", finishtime)

	var message string
	if date1.After(date2) {
		message = "開始時間より終了時間を後にしてください！"
		renderTemplate(w, "index.html", Data{Message: message})
		return
	} else if date1.Equal(date2) {
		message = "開始時間と終了時間が同じになっています"
		renderTemplate(w, "index.html", Data{Message: message})
		return
	}

	db, err := ConnectDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer db.Close()
	//データベースに挿入
	err = InsertTimeData(db, tableName, starttime, finishtime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	tmpl, err := template.ParseFiles("handler.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p := Time{Starttime: starttime, Finishtime: finishtime}
	err = tmpl.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func convertTableName(judgeDataForm string) string {
	switch judgeDataForm {
	case "default":
		tableName := "coffeetime_default"
		return tableName
	case "assignment":
		tableName := "coffeetime"
		return tableName
	}
	return ""
}

func Coffeetime(w http.ResponseWriter, r *http.Request) {
	tableName := "coffeetime"
	records, err := getUsers(tableName)
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.ParseFiles("timeSelect.html")
	if err != nil {
		log.Fatal(err)
	}

	if err := tmpl.Execute(w, records); err != nil {
		log.Fatal(err)
	}
}

func coffeetime_default(w http.ResponseWriter, r *http.Request) {
	tableName := "coffeetime_default"
	records, err := getUsers(tableName)
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err := template.ParseFiles("timeSelect.html")
	if err != nil {
		log.Fatal(err)
	}

	if err := tmpl.Execute(w, records); err != nil {
		log.Fatal(err)
	}
}

func userTimeSelect(w http.ResponseWriter, r *http.Request) {
	records, err := getUsersTime()
	if err != nil {
		log.Fatal(err)
	}

	array := changetime(records[0].Starttime, records[0].Finishtime)

	tmpl, err := template.New("user.html").Funcs(template.FuncMap{
		"first": func(array []string) template.HTML {
			if len(array) > 0 {
				return template.HTML(array[0])
			}
			return ""
		},
		"last": func(array []string) template.HTML {
			if len(array) > 0 {
				return template.HTML(array[len(array)-1])
			}
			return ""
		},
	}).ParseFiles("user.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, array); err != nil {
		log.Fatal(err)
	}

}
