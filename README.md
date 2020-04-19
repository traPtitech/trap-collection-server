## Go API Server for openapi

traPCollectionのAPI

- API version: 1.0.0
For more information, please visit [https://github.com/traPtitech/trap-collection-server](https://github.com/traPtitech/trap-collection-server)

### Mockの再生成
swaggerを書き換えたら必ず再生成してください。
プロジェクトルートで
```
# docker run -it --rm\
    -v ${PWD}:/local\
    -w /home/groovy/scripts groovy groovy /local/generate/generator.groovy generate \
    -i /local/docs/swagger/openapi.yml \
    -g CollectionCodegen \
    -t /local/generate \
    -o /local
```

### コードの書き換え
`main.go`,`README.md`,`openapi/`は書き換えないでください。
書き換えても再生成で全て消えます。
これらのファイルを書き換えたい場合は大抵`generate/`または`docs/swagger/openapi.yml`を書き換えることで対応できます。
