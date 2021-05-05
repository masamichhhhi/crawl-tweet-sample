include .env

$(eval export $(shell sed -ne 's/ *#.*$//; /./ s/=.*$$// p' .env))

.PHONY: build
# ビルド
build:
	sam build

deploy:
	sam deploy --stack-name crawl-tweet-sample --parameter-overrides AwsAccessKey=${AWS_ACCEESS_KEY} AwsSecretAccessKey=${AWS_SECRET_ACCEESS_KEY} ConsumerKey=${CONSUMER_KEY} ConsumerSecret=${CONSUMER_SECRET} AccessToken=${ACCESS_TOKEN} AccessTokenSecret=${ACCESS_TOKEN_SECRET}

# デプロイしたlambdaの削除
delete:
	aws cloudformation delete-stack --stack-name crawled-tweet-sample

# ローカルで実行
invoke:
	sam build && sam local invoke --parameter-overrides AwsAccessKey=${AWS_ACCEESS_KEY} AwsSecretAccessKey=${AWS_SECRET_ACCEESS_KEY} ConsumerKey=${CONSUMER_KEY} ConsumerSecret=${CONSUMER_SECRET} AccessToken=${ACCESS_TOKEN} AccessTokenSecret=${ACCESS_TOKEN_SECRET}
