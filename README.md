## trap-collection-server
[![codecov](https://codecov.io/gh/traPtitech/trap-collection-server/branch/main/graph/badge.svg)](https://codecov.io/gh/traPtitech/trap-collection-server)
[![](https://github.com/traPtitech/trap-collection-server/workflows/Release/badge.svg)](https://github.com/traPtitech/trap-collection-server/actions)
[![swagger](https://img.shields.io/badge/swagger-docs-brightgreen)](https://apis.trap.jp/?urls.primaryName=traP%20Collection)
[![go report](https://goreportcard.com/badge/traPtitech/trap-collection-server)](https://goreportcard.com/report/traPtitech/trap-collection-server)

traPのゲームランチャーtraP Collectionのサーバーサイドです。

### 開発環境の起動
`.env`ファイルに
```
CLIENT_ID={{traQのClientのClientID}}
CLIENT_SECRET={{traQのClientのClientSecret}}
```
のように書き、
```
$ docker compose -f docker/dev/compose.yaml up
```
をすると動きます。
