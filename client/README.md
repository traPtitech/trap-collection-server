# anke-to UI

## Project setup
```
npm install
```

### Compiles and hot-reloads for development
```
npm run serve
```

### Compiles and minifies for production
```
npm run build
```

### Run your tests
```
npm run test
```

### Lints and fixes files
```
npm run lint
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).

## URL

- `/targeted` : 自分が対象になっているアンケート一覧
- `/administrates` : 自分が管理者になっているアンケート一覧
- `/responses` : 自分の回答一覧
- `/explorer` : すべてのアンケート一覧
- `/questionnaires/:id` : questionnaireID = id のアンケートの詳細
  - 編集のときは `#edit` を末尾につける
- `/results/:id` : questionnaireID = id のアンケートの結果
- `/responses/:id` : responseID = id の回答
  - 編集のときは `#edit` を末尾につける
