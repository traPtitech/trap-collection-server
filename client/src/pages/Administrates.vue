<template>
  <div class="wrapper">
    <div class="list card content">
      <header class="card-header">
        <div class="card-header-title subtitle">
          自分が管理者になっているアンケート
        </div>
      </header>
      <div class="card-content">
        <article
          v-for="(questionnaire, index) in questionnaires"
          :key="index"
          class="post"
        >
          <div>
            <div class="questionnaire-title">
              <span
                :class="{
                  'ti-check': questionnaire.status === 'all-respond',
                  'ti-alert': questionnaire.status === 'not-all-respond'
                }"
              ></span>
              <span class="subtitle">
                <router-link
                  :to="'/questionnaires/' + questionnaire.questionnaireID"
                  >{{ questionnaire.title }}</router-link
                >
              </span>
            </div>
            <p>{{ questionnaire.description }}</p>
            <div class="media">
              <div class="media-content has-text-weight-bold columns">
                <div class="content column res-time-limit">
                  回答期限: {{ getDateStr(questionnaire.res_time_limit) }}
                </div>
                <div class="content column modified-at">
                  更新日: {{ getRelativeDateStr(questionnaire.modified_at) }}
                </div>
              </div>
            </div>
          </div>
        </article>
      </div>
    </div>
  </div>
</template>

<script>
import axios from '@/bin/axios'
import common from '@/bin/common'

export default {
  name: 'Administrates',
  components: {},
  props: {},
  data() {
    return {
      questionnaires: [],
      headers: [
        'Title',
        'Time Limit',
        'Response',
        'Modified At',
        'Results',
        'Details'
      ]
    }
  },
  computed: {},
  async created() {
    axios.get('/users/me/administrates').then(resp => {
      this.questionnaires = resp.data
      this.getStatus()
    })
  },
  methods: {
    getDateStr(str) {
      return common.getDateStr(str)
    },
    getRelativeDateStr(str) {
      return common.relativeDateStr(str)
    },
    getStatus() {
      for (let i = 0; i < this.questionnaires.length; i++) {
        if (this.questionnaires[i].all_responded) {
          // 全員回答済み
          this.setStatus(i, 'all-respond')
        } else {
          // 未回答者がいる
          this.setStatus(i, 'not-all-respond')
        }
      }
    },
    setStatus(index, newStatus) {
      let questionnaire = this.questionnaires[index]
      this.$set(questionnaire, 'status', newStatus)
      this.$set(this.questionnaires, index, questionnaire)
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.content {
  margin-left: 1.5rem;
  p {
    margin-bottom: 0.5em;
    word-break: break-all;
    line-height: 1.1em;
  }
  h4 {
    margin-bottom: 0.7em;
  }
}
.content.column {
  padding: 0;
  margin-bottom: 0;
  display: inline-block;
}
.content.column.res-time-limit {
  width: 15rem;
}
.content.column.modified-at {
  width: 10rem;
}
article.post {
  padding: 1rem;
  /* padding-bottom: 0; */
  border-bottom: 1px solid #e6eaee;
}
.columns {
  padding-top: 0;
  .media-content {
    padding-top: 1em;
  }
}
.column {
  padding-left: 0;
  .media-right {
    margin: auto;
  }
}
.questionnaire-title {
  padding-bottom: 1rem;
}
</style>
