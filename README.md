## trap-collection-server

traPのゲームランチャーtraP Collectionのサーバーサイドです。

### ディレクトリの構成
OpenAPI Generatorにより、openapiディレクトリ内にルーティング関連の関数（Bodyからパラメーターの取り出しなどを行う）、main.go、が生成されます。

### Mockの起動
ドライブの[traP Collectionのフォルダ](https://drive.trap.jp/f/399071)にある`collection-mock`内のデータを`upload`ディレクトリへ移したあと、
`.env`ファイルに
```
CLIENT_ID={{traQのClientのClientID}}
CLIENT_SECRET={{traQのClientのClientSecret}}
```
のように書き、
```
$ sudo COMPOSE_DOCKER_CLI_BUILD=1 DOCKER_BUILDKIT=1 docker-compose -f docker/mock/docker-compose.yml up
```
をすると動きます。

### コードの生成
最初にする必要があります。
swaggerの変更をしたときにも行ってください。
groovyで本家OpenAPI Generatorを使っている関係で実行にそれなりに時間がかかります。
```
# docker run -it --rm \
    -v $PWD:/local \
    -w /home/groovy/scripts \
    groovy:3.0.2 \
    groovy /local/generate/generator.groovy generate \
    -i /local/docs/swagger/openapi.yml \
    -g CollectionCodegen \
    -t /local/generate \
    -o /local
```

### コードの書き換え
`main.go`,`README.md`,`openapi/`は書き換えないでください。
書き換えても再生成で全て消えます。
これらのファイルを書き換えたい場合は大抵`generate/`または`docs/swagger/openapi.yml`を書き換えることで対応できます。
