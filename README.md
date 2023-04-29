# CoffeeTaker
  ## アプリの説明
  シンプルなコーヒー予約アプリです.<br>
  このアプリでは管理者（コーヒーを作る人）が１日の中でコーヒーを出せる時間帯を指定すると、ユーザー（コーヒーを作ってもらう人）がその時間帯の中からコーヒーを作ってほしい時間を指定できるシステムになっています。<br>
  ユーザーが時間を指定すると管理者の **Line** に通知が届きます。<br>
  デフォルトの時間を設定することで毎日、時間帯を設定する手間を省くことができます。<br>
  また「デフォルトの時間」「指定の時間」の両方を設定している場合には、指定の時間を優先してユーザーに表示されます。<br>
  管理者が指定した「デフォルトの時間」「指定の時間」を確認したい場合は管理者画面からそれぞれの履歴に飛び、確認することができます。
  
  ## デモページ
  - PC
 

https://user-images.githubusercontent.com/95427365/235276721-a5336f09-01de-4e0d-adc5-c5845cd15c7c.mp4


  ## 注意点
  時間を指定する場合はモバイル端末で行うとエラーが起きる可能性があります。
  
  ## 作ることになった経緯
  兄が作業中、就寝中、お出かけ中でコーヒーを作って欲しいと頼みづらいことがありました。<br>
  そこでいつでも簡単に頼むことが出来ないかと思いこのアプリを作る事にしました。
  
  ## こだわった点
  ユーザが指定時間を選ぶ際に毎分刻みになってしまうと、選択する量が増え選ぶのが大変になるので「管理者」 「ユーザー」共に３０分刻みで選べるように設定しました。<br>
  お家で使う用なので他の方法より楽になるよう、必要最低限の機能に絞りました。<br>
  今回はサーバーも自身で組もうと思い、使用言語をGo言語にしました。<br>
  またコーヒーを頼んだ際に管理者に届く通知を出来るだけ本人の目に届きやすくしたいと思い、Lineで通知を送る事にしました。
  
  ## つまづいたところ
  Go言語ではリクエストとレスポンスがURLでやり取りを行っていたので、テンプレート内でのaタグのhref属性をどのように設定すれば色んな端末で使えるように出来るのかがわからなかったこと。<br>
  Go言語ではサーバーを自身で実装するため、ハンドラとテンプレート、リクエストとレンスポンスを理解してプログラムを組んでいく必要があったこと。<br>
  データベースから取得した値を加工する作業をどうすればシンプルに出来るかが迷ったこと。
  
  
  ## 準備(Mac)
  - postgresqlのインストール
  - データベースの作成
      - データベース名はなんでも良い(後にプログラムに入れる値ではあります)<br>
      - データベース作成コマンド
      ``` CREATE DATABASE データベース名 ```
  - テーブルの作成(テーブル名をcoffeetimeとcoffeetime_defaultdeで行う)
    - テーブル作成コマンド<br>
    ```postgresql
    CREATE TABLE 各テーブル名 (
	  id SERIAL PRIMARY KEY,
	  created_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() AT TIME ZONE 'Asia/Tokyo'),
	  starttime varchar(30),
	  finishtime varchar(30));
    ```
  - lineのアクセストークンの取得→[アクセストークン取得方法はこちらのサイトを参考にしています。](https://firestorage.jp/business/line-notify/)
  - contentsディレクトリに.envファイルを作成
  - 環境変数のHOST(データベースのホスト名)、PORT(データベースのポート名)、USER(データベースのユーザ名)、PASSWORD(データベースのパスワード)、DBNAME(データベース名)、ACCESSTOKEN(後にLineで取得するアクセストークン)
  - githubからpostgresqlと環境変数を使うためのライブラリを取得する。
    (postgresql)
    ```
    go get github.com/joho/godotenv
    ```
    (環境変数)
    ```
    go get github.com/lib/pq
    ```
  
  ## 準備(windows)
   - postgresqlのインストール
   - データベースの作成
       - データベース名はなんでも良い(後にプログラムに入れる値ではあります)<br>
       - データベース作成コマンド
       ``` CREATE DATABASE データベース名 ```
  - テーブルの作成(テーブル名をcoffeetimeとcoffeetime_defaultdeで行う)
    - テーブル作成コマンド<br>
    ```postgresql
    CREATE TABLE 各テーブル名 (
	  id SERIAL PRIMARY KEY,
	  created_at TIMESTAMP WITH TIME ZONE DEFAULT (NOW() AT TIME ZONE 'Asia/Tokyo'),
	  starttime varchar(30),
	  finishtime varchar(30));
    ```
  - lineのアクセストークンの取得→[アクセストークン取得方法はこちらのサイトを参考にしています。](https://firestorage.jp/business/line-notify/)
  - contentsディレクトリに.envファイルを作成
  - 環境変数のHOST(データベースのホスト名)、PORT(データベースのポート名)、USER(データベースのユーザ名)、PASSWORD(データベースのパスワード)、DBNAME(データベース名)、ACCESSTOKEN(後にLineで取得するアクセストークン)
  - modファイルの作成
    ```
    go mod init 
    ```
  - 不足しているモジュールの追加(go install でも試しましたが私の環境ではimportが使える状態にはなりませんでした。)
    ```
    go mod tidy
    ```
  
  ## 使用方法
  - contentsディレクトリで
    ```
    go run main.go 
    ```
    を実行する
  - **アプリケーション“main”へのネットワーク受信接続を許可しますか?** と表示されるので**許可**を選択する
  - 実行パソコンで接続する
  	- URLに```http://localhost:8080/``` を入力し接続する。
  - 自宅の携帯で接続する(Wifiを通して繋げられるようになります。)
  	- 実行パソコンのローカルループバックアドレスを調べる
  	- URLに```http://localhost:8080/```の```localhost```の部分を調べたローカルループバックアドレスに変える
  	- 例　→ ```http://120.0.0.1:8080/```
  	
  
