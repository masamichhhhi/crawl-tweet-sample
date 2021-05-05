package main

import (
	"log"
	"net/url"
	"os"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/guregu/dynamo"
)

// ①dynamoDBに入れたい情報のstructの定義
// User つぶやいたユーザ情報
type User struct {
	ID         int64  `dynamo:"id"`
	Name       string `dynamo:"name"`
	ScreenName string `dynamo:"screen_name"`
}

// Tweet 参加を募集するツイート
type Tweet struct {
	ID        string `dynamo:"tweet_id"` //パーティションキー
	FullText  string `dynamo:"full_text"`
	TweetedAt int64  `dynamo:"tweeted_at"` //dynamodbでソート出来るようにUNIX時間
	ExpiredAt int64  `dynamo:"expired_at"`
	User      User   `dynamo:"user"`
}

// NewTweet Tweetのメンバを初期化する関数
func NewTweet() Tweet {
	tweet := Tweet{}
	tweet.ID = ""
	tweet.FullText = ""
	tweet.TweetedAt = 0
	tweet.ExpiredAt = 0
	tweet.User = User{}
	return tweet
}

func crawlTweets() {
	// ②Twitter APIと AWSの認証
	const tableName = "crawled_tweet"

	creds := credentials.NewStaticCredentials(os.Getenv("AWS_ACCEESS_KEY"), os.Getenv("AWS_SECRET_ACCEESS_KEY"), "") //第３引数はtoken

	sess, _ := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String("ap-northeast-1")},
	)

	db := dynamo.New(sess)
	table := db.Table(tableName)

	anaconda.SetConsumerKey(os.Getenv("CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("CONSUMER_SECRET"))

	api := anaconda.NewTwitterApi(os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_TOKEN_SECRET"))

	// 検索するオプション
	v := url.Values{}
	v.Set("count", "30")

	searchResult, _ := api.GetSearch("#Golang", v)

	// ③DynamoDBにツイートを保存
	// 文字列→日付に変換するレイアウト
	var layout = "Mon Jan 2 15:04:05 +0000 2006"

	// 期限を一日後に設定する
	expiredAt := time.Now().AddDate(0, 0, 1).Unix()
	for _, tweet := range searchResult.Statuses {

		newTweet := NewTweet()
		tweetedTime, _ := time.Parse(layout, tweet.CreatedAt)

		if tweet.RetweetedStatus == nil {
			newTweet.ID = tweet.IdStr
			newTweet.FullText = tweet.FullText
			newTweet.TweetedAt = tweetedTime.Unix()
			newTweet.ExpiredAt = expiredAt
			newTweet.User = User{
				tweet.User.Id,
				tweet.User.Name,
				tweet.User.ScreenName,
			}

		} else {

			newTweet.ID = tweet.RetweetedStatus.IdStr
			newTweet.FullText = tweet.RetweetedStatus.FullText
			newTweet.TweetedAt = tweetedTime.Unix()
			newTweet.ExpiredAt = expiredAt
			newTweet.User = User{
				tweet.RetweetedStatus.User.Id,
				tweet.RetweetedStatus.User.Name,
				tweet.RetweetedStatus.User.ScreenName,
			}
		}

		if err := table.Put(newTweet).If("attribute_not_exists(tweet_id)").Run(); err != nil {
			log.Println(err.Error())
		} else {
			log.Println("Success！")
		}

	}

}

func main() {
	// ラムダ実行
	lambda.Start(crawlTweets)
}
