<template>
  <div>
    <div v-if="canViewResults" class="details is-fullheight">
      <div class="tabs is-centered">
        <router-link id="return-button" :to="summaryProps.titleLink">
          <span class="ti-arrow-left"></span>
        </router-link>
        <ul>
          <li
            v-for="(tab, index) in detailTabs"
            :key="index"
            class="tab"
            :class="{ 'is-active': selectedTab === tab }"
          >
            <router-link :to="getTabLink(tab)">{{ tab }}</router-link>
          </li>
        </ul>
      </div>
      <information-summary :information="summaryProps"></information-summary>
      <component
        :is="currentTabComponent"
        class="details-child is-fullheight"
        :name="currentTabComponent"
        :results="results"
        :information="information"
        :questions="questions"
        :question-data="questionData"
        :response-data="responseData"
        @get-results="getResults"
      ></component>
    </div>

    <div
      v-if="information.administrators && !canViewResults"
      class="message is-danger"
    >
      <p class="message-body error-message">結果を閲覧する権限がありません</p>
    </div>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import axios from '@/bin/axios'
import common from '@/bin/common'
import Individual from '@/components/Results/Individual'
import Spreadsheet from '@/components/Results/Spreadsheet'
import InformationSummary from '@/components/Information/InformationSummary'

export default {
  name: 'Results',
  components: {
    individual: Individual,
    spreadsheet: Spreadsheet,
    'information-summary': InformationSummary
  },
  props: {},
  data() {
    return {
      results: [],
      questions: [],
      questionData: [],
      responseData: {},
      information: {},
      hasResponded: false,
      detailTabs: ['Spreadsheet', 'Individual']
    }
  },
  computed: {
    ...mapGetters(['getMyTraqId']),
    questionnaireId() {
      return this.$route.params.id
    },
    administrates() {
      if (!this.information.administrators) {
        return undefined
      }
      return common.administrates(
        this.information.administrators,
        this.getMyTraqId
      )
    },
    canViewResults() {
      return common.canViewResults(
        this.information,
        this.administrates,
        this.hasResponded
      )
    },
    currentTabComponent() {
      switch (this.selectedTab) {
        case 'Spreadsheet':
          return 'spreadsheet'
        case 'Individual':
          return 'individual'
        default:
          console.error('unexpected selectedTab')
          return ''
      }
    },
    selectedTab() {
      return this.$route.query.tab && this.$route.query.tab === 'individual'
        ? 'Individual'
        : 'Spreadsheet'
    },
    currentPage() {
      if (this.$route.query.tab === 'individual') {
        return this.$route.query.page ? Number(this.$route.query.page) : 1
      } else {
        return undefined
      }
    },
    summaryProps() {
      let ret = {
        title: this.information.title,
        titleLink: '/questionnaires/' + this.questionnaireId
      }
      if (this.selectedTab === 'Individual') {
        ret.responseDetails = {
          timeLabel: '回答日時',
          time: this.responseData.submittedAt,
          respondent: this.responseData.traqId
        }
      }
      return ret
    }
  },
  watch: {
    $route: function(newRoute) {
      if (newRoute.query.tab === 'individual') {
        this.setResponseData()
        this.setResponsesToQuestions()
      }
    }
  },
  async created() {
    this.getInformation()
      .then(this.getMyResponses)
      .then(() => {
        if (this.canViewResults) {
          this.getResults('')
            .then(this.getQuestions)
            .then(() => {
              if (this.$route.query.tab === 'individual') {
                this.setResponseData()
                this.setResponsesToQuestions()
              }
            })
        }
      })
  },
  methods: {
    getDateStr: common.getDateStr,
    async getResults(query) {
      return axios.get('/results/' + this.questionnaireId + query).then(res => {
        this.results = []
        res.data.forEach(data => {
          this.results.push({
            modifiedAt: this.getDateStr(data.modified_at),
            responseId: data.responseID,
            responseBody: data.response_body,
            submittedAt: this.getDateStr(data.submitted_at),
            traqId: data.traqID
          })
        })
      })
    },
    getQuestions() {
      this.questions = []
      this.questionData = []
      return axios
        .get('/questionnaires/' + this.questionnaireId + '/questions')
        .then(res => {
          for (const question of res.data) {
            this.questions.push(question.body)
            this.questionData.push(common.convertDataToQuestion(question))
          }
        })
    },
    getInformation() {
      return axios.get('/questionnaires/' + this.questionnaireId).then(res => {
        this.information = res.data
      })
    },
    getMyResponses() {
      return axios
        .get('/users/me/responses/' + this.questionnaireId)
        .then(res => {
          if (res.data.length > 0) {
            this.hasResponded = true
          }
        })
    },
    getTabLink(tab) {
      let ret = {
        name: 'Results',
        params: { id: this.$route.params.id },
        query: {}
      }
      if (tab === 'Individual') {
        ret.query.tab = 'individual'
      } else {
        ret.query.tab = 'spreadsheet'
      }
      return ret
    },
    setResponseData() {
      this.responseData = this.results[this.currentPage - 1]
      let newBody = {}
      this.responseData.responseBody.forEach(data => {
        newBody[data.questionID] = data
      })
      this.responseData.body = newBody
    },
    setResponsesToQuestions() {
      const questions = Object.assign([], this.questionData)
      questions.forEach((question, index) => {
        this.$set(
          this.questionData,
          index,
          common.setResponseToQuestion(
            question,
            this.responseData.body[question.questionId]
          )
        )
      })
    },
    setResults(results) {
      this.results = results
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.error-message {
  font-size: 1rem;
  margin: 1rem;
}
</style>
