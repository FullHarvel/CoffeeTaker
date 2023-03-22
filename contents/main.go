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

// github.com/lib/pqはgo mod init github.com/lib/pqとgo get github.com/lib/pq"をしたらエラーが出なくなった
// グローバル変数の定義
var host string
var port string
var user string
var password string
var dbname string
var accessToken string

// 環境変数のセット
func setEnv() {
	// .envファイルから環境変数を読み込む
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	// 環境変数をグローバル変数に格納する
	//port変数もstring型になっているので使う場合はint型として変数に入れて使う必要がある
	host = os.Getenv("HOST")
	port = os.Getenv("PORT")
	user = os.Getenv("USER")
	password = os.Getenv("PASSWORD")
	dbname = os.Getenv("DBNAME")
	accessToken = os.Getenv("ACCEESSTOKEN")
}

// メインの関数
func main() {
	setEnv()

	//cssとかjsの静的ファイルをstaticファイルの中に入れると読み込めるようにしている
	//これをする場合htmlのcssの設定でhrefをhref="/static/style.css"とかにしないといけない
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static/"))))

	//ルーターを設定
	//最初のページ
	http.HandleFunc("/", mainPage)
	//管理者の予約時間を指定するページ
	http.HandleFunc("/index", index)
	//予約時間を指定した時に出るページ
	http.HandleFunc("/handler", handler)
	//予約の履歴を参照するページ
	http.HandleFunc("/getCoffeeTime", Coffeetime)
	http.HandleFunc("/getCoffeeTimeDefault", coffeetime_default)
	//ユーザのページ
	http.HandleFunc("/user", userTimeSelect)
	http.HandleFunc("/userConfirm", userConfirm)

	fmt.Println("start")
	//サーバーを立てる
	//マルチプレクサみたいなのを起動させてるのかな？
	http.ListenAndServe("localhost:8080", nil)
}

// lineに通知を送る
func sendLineMessage(message string) {

	lineAccessToken := accessToken
	//今日の日にちを取得
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

//ユーザー側の処理

// 今変な名前にしてる
// データベースにある予約開始時間と予約終了時間を取得する
func getUsersTime() ([]Record, error) {
	db, err := ConnectDB()
	defer db.Close()

	var records []Record
	var r Record
	//管理者側で予約の指定があるかどうかを確認する
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

// changetimeはこっちを採用
func changetime(starttime string, finishtime string) []string {
	//管理者の設定した時間をユーザーが選べる様にする(makeTimeId関数用)
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

	//上記の開始のインデックス、終了のインデックスを取得する
	startIndex := makeTimeId(allTimes, starttime)
	finishIndex := makeTimeId(allTimes, finishtime)

	result := getTime(allTimes, startIndex, finishIndex)

	return result
}

// 時間に対応したIDを返す
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

// ユーザが予約時間を選んだ後のユーザに知らせる画面
func userConfirm(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //フォームデータを使える形にする
	//ここはname := r.FormValue("name")で受け取ることもできるみたい
	message := r.Form.Get("message")
	sendLineMessage(message)
	http.ServeFile(w, r, "user_conf.html")

}

// 管理者側でデフォルトの値か決めた値を使うのかを決めるために今日の日付の値があるのかをデータベースにあるのかを見るため
type IsTodayData struct {
	Today      time.Time
	Starttime  string
	Finishtime string
}

// 今日の日付のデータがcoffeetimeのテーブルのcreated_tableに入ってるのかどうかを調べる
// デフォルトの予約時間か指定した予約時間かどちらを使うのか判定するためにある
// 指定した時間を使う場合はそのまま取ってきたstarttimeとfinishtimemを使う
func isTodayData() (bool, string, string) {
	//データベース接続
	db, err := ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 今日の日本時間を取得
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Now().In(jst).Format("2006-01-02")

	//データベースから追加した日付を取得
	rows, err := db.Query("SELECT created_at, starttime, finishtime  FROM coffeetime order by id desc limit 1")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	//取り出してきた値を日付だけにしている
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

//管理者側のやつ
// データベース関係
// データベースに入れる値
// ポート番号

type Data struct {
	Message string
}

// テンプレートを実行する関数（開始時間と終了時間のバリデーションに使ってる）
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

// データベース接続
// データベースはこのコマンドで作った（デフォルトの時間設定のやつ）
// CREATE TABLE coffeetime_default (
//
//	id SERIAL PRIMARY KEY,
//	created_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() AT TIME ZONE 'Asia/Tokyo'),
//	starttime varchar(30),
//	finishtime varchar(30)
//
// );
func ConnectDB() (*sql.DB, error) {
	//string型になっているポート番号をintの型に変更する
	port, _ := strconv.Atoi(port)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	//本当にサーバーと接続ができているか確認するもので
	//Pingを送信しサーバーが応答しなかった場合は接続できていないとみなす
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// データベースに今日のコーヒーの予約開始時間と予約終了時間を入れる
func InsertTimeData(db *sql.DB, tableName string, data1 string, data2 string) error {
	// SQL文の作成
	sqlStatement := `
        INSERT INTO ` + tableName + `(starttime,finishtime)
        VALUES ($1, $2)`

	// パラメータの設定
	//index.htmlのpostで取ってきた予約開始、予約終了時間を入れている
	starttime := data1
	finishtime := data2

	// SQL文の実行
	//db.Exec(データベースを挿入するクエリ,)第二引数からh入れたい値を数の分だけ入れるとクエリの$1とか$2とかに順番に入っていく
	_, err := db.Exec(sqlStatement, starttime, finishtime)
	fmt.Println(err)
	if err != nil {
		return err
	}
	return nil
}

// データベースの値を入れる構造体
// データベースの値をテンプレートに入れるときに使う
type Record struct {
	Today      time.Time
	Starttime  string
	Finishtime string
}

// データベースにある予約開始時間と予約終了時間の履歴を管理者側が見る
// 引数にどのテーブルに入れるかを指定する
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

// 最初の画面(単純にhtmlを返す)
// 管理者画面に飛ぶかユーザ画面に飛ぶか決める画面
func mainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "main.html")
}

// 今の所ここから自分の指定の履歴画面とかが見れたらいいかなと思う
func index(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", Data{Message: ""})
	return
}

// ハンドラーの設定
// データベースから取得してきた値を入れる
type Time struct {
	Starttime  string
	Finishtime string
}

// 管理者画面から入力された値をデータベースに入れてからその値をhandler.htmlの画面に表示する
func handler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //フォームデータを使える形にする
	//ここはname := r.FormValue("name")で受け取ることもできるみたい
	judgeDataForm := r.Form.Get("judgeDataForm")

	starttime := r.Form.Get("starttime")
	finishtime := r.Form.Get("finishtime")
	tableName := convertTableName(judgeDataForm)

	date1, _ := time.Parse("15:04", starttime)
	date2, _ := time.Parse("15:04", finishtime)

	//開始時関より終了時間の方が早かった時のバリデーション
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

	//データベースにはstarttime（開始時間）とfinishtime（終了時間）を入れる
	//データベース接続
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

	//テンプレートを作成
	//データベースを入れた後にデータベースに値を入れたっていうことをユーザに伝える画面
	tmpl, err := template.ParseFiles("handler.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//テンプレートに入れる値は構造体に値を当てはめ込んだやつのkeyの方がテンプレートの値と一致している必要がある
	p := Time{Starttime: starttime, Finishtime: finishtime}
	err = tmpl.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// 管理者のフォーム画面からデフォルトで値を入れるのか指定で入れるのかを判断してテーブル名を変更するやつ
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

// データベースから取得してきた値を表示
func Coffeetime(w http.ResponseWriter, r *http.Request) {
	//データベースのテーブルをcoffeetimeテーブルに設定
	tableName := "coffeetime"
	//データベースの値を取得
	records, err := getUsers(tableName)
	if err != nil {
		log.Fatal(err)
	}

	//テンプレート作成
	tmpl, err := template.ParseFiles("timeSelect.html")
	if err != nil {
		log.Fatal(err)
	}

	//テンプレートに値を当てはめる
	if err := tmpl.Execute(w, records); err != nil {
		log.Fatal(err)
	}
}

// データベースから取得してきた値を表示
func coffeetime_default(w http.ResponseWriter, r *http.Request) {
	//データベースのテーブルをcoffeetimeテーブルに設定
	tableName := "coffeetime_default"
	//データベースの値を取得
	records, err := getUsers(tableName)
	if err != nil {
		log.Fatal(err)
	}

	//テンプレート作成
	tmpl, err := template.ParseFiles("timeSelect.html")
	if err != nil {
		log.Fatal(err)
	}

	//テンプレートに値を当てはめる
	if err := tmpl.Execute(w, records); err != nil {
		log.Fatal(err)
	}
}

// ユーザ画面を作成
func userTimeSelect(w http.ResponseWriter, r *http.Request) {
	//データベースから一番最新の値を取り出してくる
	records, err := getUsersTime()
	if err != nil {
		log.Fatal(err)
	}

	//予約の開始時間から終了時間まで３０分刻みにして配列に入れる
	array := changetime(records[0].Starttime, records[0].Finishtime)

	//テンプレート作成
	//それとテンプレートに当てはめるために関数を設定もしてる
	//テンプレートでは{{first .}}とか{{last .}}で関数を使用する事ができる
	//Newの引数にはにはテンプレートを一意にするためにあり、ファイル名やディレクトリ名を入れる
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

	//テンプレートに値を当てはめる
	if err := tmpl.Execute(w, array); err != nil {
		log.Fatal(err)
	}

}

//今日の日付の設定がデータベース上にあったら絞るやつをデータベースからとってきてなかったらデフォルトのやつを取ってくるって言うのもしないといけない
//lineを送る機能、時間指定の最後のデータの削除機能,管理者側のみログイン機能（lineの通知設定のため）、ユーザの登録した時間の履歴（最新のみ）
