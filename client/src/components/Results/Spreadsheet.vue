<template>
  <div class="wrapper">
    <div class="card">
      <div class="tabs">
        <ul>
          <li
            v-for="tab in tableFormTabs"
            :key="tab"
            class="tab"
            :class="{ 'is-active': tableForm === tab }"
          >
            <a @click="tableForm = tab">{{ tab }}</a>
          </li>
        </ul>
        <button
          v-if="canDownload"
          class="button download"
          @click="downloadTable"
        >
          <span class="ti-download"></span>
        </button>
      </div>
      <div class="scroll-view">
        <!-- table view -->
        <table v-show="tableForm === 'view'" class="table is-striped">
          <thead>
            <tr>
              <th
                v-for="(header, index) in headerLabels.concat(questions)"
                :key="index"
                :class="{ active: sorted == index + 1 || sorted == -1 - index }"
                @click="sort(index + 1)"
              >
                {{ header }}
                <span
                  class="arrow"
                  :class="sorted !== index + 1 ? 'asc' : 'dsc'"
                ></span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(result, index) in results" :key="index">
              <td class="table-item-traqid">{{ result.traqId }}</td>
              <td class="table-item-time">{{ result.submittedAt }}</td>
              <td
                v-for="response in result.responseBody"
                :key="response.responseId"
              >
                {{ getResponse(response) }}
              </td>
            </tr>
          </tbody>
        </table>

        <!-- markdown view -->
        <textarea
          v-show="tableForm === 'markdown'"
          class="textarea"
          :value="markdownTable"
          :rows="results.length + 3"
          readonly
        ></textarea>

        <!-- csv view -->
        <textarea
          v-show="tableForm === 'csv'"
          class="textarea"
          :value="csvTable"
          :rows="results.length + 2"
          readonly
        ></textarea>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Spreadsheet',
  components: {},
  props: {
    results: {
      type: Array,
      required: true
    },
    questions: {
      type: Array,
      required: true
    }
  },
  data() {
    return {
      headers: [
        { name: 'traqId', label: 'traQID' },
        { name: 'submittedAt', label: '回答日時' }
      ],
      sorted: '',
      downloadLabel: 'CSV形式でダウンロード',
      tableForm: 'view', // 'view', 'markdown', 'csv'
      tableFormTabs: ['view', 'markdown', 'csv']
    }
  },
  computed: {
    questionnaireId() {
      return this.$route.params.id
    },
    headerLabels() {
      let ret = []
      this.headers.forEach(header => {
        ret.push(header.label)
      })
      return ret
    },
    markdownTable() {
      // results の表を markdown 形式にしたものを返す
      let ret = ''

      // 項目の行
      ret += this.arrayToMarkdown(this.headerLabels.concat(this.questions))

      // 2行目
      ret += '|'
      for (let i = 0; i < this.results.length + this.headers.length; i++) {
        ret += ' - |'
      }
      ret += '\n'

      // 各回答の行
      for (let i = 0; i < this.results.length; i++) {
        let arr = []
        const result = this.results[i]
        this.headers.forEach(header => {
          arr.push(result[header.name])
        })
        result.responseBody.forEach(body => {
          arr.push(this.getResponse(body))
        })

        ret += this.arrayToMarkdown(arr)
      }

      ret = ret.slice(0, -1) // 末尾の改行を削除

      return ret
    },
    csvTable() {
      let csv = '\ufeff'
      this.headerLabels.concat(this.questions).forEach(header => {
        if (csv !== '\ufeff') {
          csv += ','
        }
        csv += '"' + header + '"'
      })
      csv += '\n'
      this.results.forEach(result => {
        csv += result.traqId + ',' + result.submittedAt
        result.responseBody.forEach(response => {
          csv += ',' + '"' + this.getResponse(response) + '"'
        })
        csv += '\n'
      })
      return csv
    },
    canDownload() {
      return this.tableForm === 'markdown' || this.tableForm === 'csv'
    }
  },
  mounted() {},
  methods: {
    getResponse(body) {
      let ret = ''
      switch (body.question_type) {
        case 'MultipleChoice':
        case 'Checkbox':
        case 'Dropdown':
          body.option_response.forEach(response => {
            if (ret !== '') {
              ret += ', '
            }
            ret += response
          })
          return ret
        default:
          return body.response
      }
    },
    downloadTable() {
      if (!this.canDownload) return
      let form = {}
      switch (this.tableForm) {
        case 'markdown':
          form = { type: 'text/markdown', ext: '.md', data: this.markdownTable }
          break
        case 'csv':
          form = { type: 'text/csv', ext: '.csv', data: this.csvTable }
          break
      }
      const blob = new Blob([form.data], { type: form.type })
      let link = document.createElement('a')
      link.href = window.URL.createObjectURL(blob)
      link.download = 'Result' + form.ext
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    },
    sort(index) {
      let query = ''
      if (this.sorted !== index) {
        query += '-'
        this.sorted = index
      } else {
        this.sorted = -index
      }
      switch (index) {
        case 1:
          query += 'traqid'
          break
        case 2:
          query += 'submitted_at'
          break
        default:
          query += index - 2
      }
      this.$emit('get-results', '?sort=' + query)
    },
    arrayToMarkdown(arr) {
      // 配列を受け取ると、その配列1行分のmarkdownを返す
      let ret = '|'
      arr.forEach(val => {
        ret += ' ' + val + ' |'
      })
      ret += '\n'
      return ret
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
th,
td {
  vertical-align: middle;
  font-size: 0.9em;
  min-width: 10em;
}

th.active {
  background-color: #fafafa;
}

.arrow {
  opacity: 1;
  color: black;
  display: inline-block;
  vertical-align: middle;
  margin: auto;
  opacity: 1;
}

.arrow.asc {
  border-left: 4px solid transparent;
  border-right: 4px solid transparent;
  border-bottom: 4px solid black;
}

.arrow.dsc {
  border-left: 4px solid transparent;
  border-right: 4px solid transparent;
  border-top: 4px solid black;
}

.download {
  margin: auto 0.5rem;
}

.card {
  overflow-x: unset;
}

.scroll-view {
  overflow-x: auto;
}

.button {
  display: block;
}
</style>
