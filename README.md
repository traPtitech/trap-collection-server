## Go API Server for openapi

traPCollectionのAPI

- API version: 1.0.0
For more information, please visit [https://github.com/traPtitech/trap-collection-server](https://github.com/traPtitech/trap-collection-server)

### Mockの起動
ドライブの[traP Collectionのフォルダ](https://drive.trap.jp/f/399071)にある`collection-mock`内のデータを`upload`ディレクトリへ移し、
`.env`ファイルに適当なtraQのClientの`CLIENT_ID,CLIENT_SECRET`の値を
```
CLIENT_ID={{traQのClientのClient ID}}
CLIENT_SECRET={{traQのClientのClient Secret}}
```
のように書いた後、
```
$ sh mockgen.sh
```
で動きます。

### コードの生成
最初にする必要があります。
swaggerの変更をしたときにも行ってください。
groovyで本家OpenAPI Generatorを使っている関係で実行にそれなりに時間がかかります。
```
# docker run -it --rm \
    -v $PWD:/local \
    -v grapes-cache:/home/groovy/.groovy/grapes \
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
