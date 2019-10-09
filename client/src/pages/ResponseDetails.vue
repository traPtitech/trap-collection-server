<template>
  <div>
    <top-bar-message :message="message"></top-bar-message>

    <div
      v-if="!isEditing || (information.res_time_limit && !timeLimitExceeded)"
      class="is-fullheight details"
    >
      <div class="tabs is-centered">
        <router-link id="return-button" :to="titleLink">
          <span class="ti-arrow-left"></span>
        </router-link>
        <ul></ul>
        <a
          v-if="
            !isNewResponse && information.res_time_limit && !timeLimitExceeded
          "
          id="edit-button"
          :class="{ 'is-editing': isEditing }"
          @click.prevent="isEditing = !isEditing"
        >
          <span class="ti-check-box"></span>
        </a>
      </div>
      <div
        :class="{ 'is-editing has-navbar-fixed-bottom': isEditing }"
        class="is-fullheight details-child"
      >
        <information-summary :information="summaryProps"></information-summary>
        <questions
          :edit-mode="isEditing ? 'response' : undefined"
          :questions-props="questions"
          :input-errors="inputErrors"
        ></questions>
      </div>
      <edit-nav-bar v-if="isEditing">
        <button
          :disabled="!submitOk || isSubmitting"
          class="button is-medium send-button"
          @click="submitResponse"
        >
          <span class="ti-check"></span>
          <span>送信</span>
        </button>
        <button
          :disabled="isSaving"
          class="button is-medium save-button"
          @click="saveResponse"
        >
          <span class="ti-save"></span>
        </button>
        <button class="button is-medium cancel-button" @click="abortEditing">
          <span class="ti-close"></span>
        </button>
      </edit-nav-bar>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import { mapGetters } from 'vuex'
import router from '@/router'
import common from '@/bin/common'
import Questions from '@/components/Questions/Questions'
import EditNavBar from '@/components/Utils/EditNavBar'
import TopBarMessage from '@/components/Utils/TopBarMessage'
import InformationSummary from '@/components/Information/InformationSummary'

export default {
  name: 'ResponseDetails',
  components: {
    questions: Questions,
    'edit-nav-bar': EditNavBar,
    'information-summary': InformationSummary,
    'top-bar-message': TopBarMessage
  },
  props: {
    isNewResponse: {
      type: Boolean,
      required: false
    }
  },
  data() {
    return {
      questions: [],
      information: {},
      responseData: {},
      message: {
        showMessage: false
      },
      isSubmitting: false,
      isSaving: false
    }
  },
  computed: {
    ...mapGetters(['getMyTraqId']),
    responseId() {
      return this.isNewResponse ? undefined : Number(this.$route.params.id)
    },
    questionnaireId() {
      if (this.isNewResponse) {
        return Number(this.$route.params.questionnaireId)
      } else if (!this.responseData) {
        return undefined
      } else {
        return this.responseData.questionnaireID
      }
    },
    isEditing: {
      get: function() {
        if (this.isNewResponse || this.$route.hash === '#edit') {
          return true
        }
        return false
      },
      set: function(newBool) {
        // newBool : 閲覧 -> 編集
        // !newBool : 編集 -> 閲覧
        const newRoute = {
          name: 'ResponseDetails',
          params: { id: this.responseId },
          hash: newBool ? '#edit' : undefined
        }
        router.push(newRoute)
      }
    },
    submitOk() {
      // 入力内容に不備がないかどうか
      for (const error of Object.keys(this.inputErrors)) {
        if (this.inputErrors[error].isError) {
          return false
        }
      }
      return true
    },
    timeLimitExceeded() {
      // 回答期限を過ぎていた場合はtrueを返す
      return (
        this.information.res_time_limit &&
        new Date(this.information.res_time_limit).getTime() <
          new Date().getTime()
      )
    },
    titleLink() {
      return '/questionnaires/' + this.questionnaireId
    },
    responseIconClass() {
      if (this.isNewResponse) {
        return undefined
      }
      switch (this.responseData.submitted_at) {
        case 'NULL':
          return 'ti-save'
        default:
          return 'ti-check'
      }
    },
    summaryProps() {
      const ret = {
        title: this.information.title,
        titleLink: this.titleLink,
        description: this.information.description,
        // timeLimit: this.getTimeLimitStr(this.information.res_time_limit),
        responseIconClass: this.responseIconClass,
        responseDetails: {
          timeLabel: '更新日時',
          time: this.getDateStr(this.responseData.modified_at),
          respondent: this.getMyTraqId
        }
      }
      return ret
    },
    inputErrors() {
      let ret = {}
      for (const question of this.questions) {
        ret[question.questionId] = {
          isError: question.isRequired && !this.hasAnswered(question),
          message: 'この質問は回答必須です'
        }
      }
      return ret
    }
  },
  watch: {
    $route: function(newRoute, oldRoute) {
      if (newRoute.params.id !== oldRoute.params.id) {
        this.resetMessage()
        this.getResponseData()
          .then(this.getQuestionnaireData)
          .then(this.getQuestions)
          .then(this.setResponsesToQuestions)
      }
    }
  },
  async created() {
    if (this.isNewResponse) {
      this.getQuestions()
      this.getInformation()
    } else {
      this.getResponseData()
        .then(this.getInformation)
        .then(this.getQuestions)
        .then(this.setResponsesToQuestions)
    }
  },
  methods: {
    alertNetworkError: common.alertNetworkError,
    getDateStr: common.getDateStr,
    getInformation() {
      return axios
        .get('/admin/questionnaires/' + this.questionnaireId)
        .then(res => {
          this.information = res.data
          if (this.timeLimitExceeded && this.isEditing) {
            this.message = {
              body: '回答期限が過ぎています',
              color: 'red',
              showMessage: true
            }
          }
        })
    },
    getResponseData() {
      return axios.get('/trap/responses/' + this.responseId).then(res => {
        this.responseData = res.data

        // questionIdをキーにしてresponseData.body の各要素をとれるようにする
        let newBody = {}
        this.responseData.body.forEach(data => {
          newBody[data.questionID] = data
        })
        this.responseData.body = newBody
      })
    },
    getQuestions() {
      this.questions = []
      return axios
        .get('/questionnaires/' + this.questionnaireId + '/questions')
        .then(res => {
          // convertDataToQuestion を通したものを this.questions に保存
          res.data.forEach(data => {
            this.questions.push(common.convertDataToQuestion(data))
          })
        })
    },
    setResponsesToQuestions() {
      // 各質問に対して、該当する回答の情報を this.questions に入れる
      this.questions.forEach((question, index) => {
        this.$set(
          this.questions,
          index,
          common.setResponseToQuestion(
            question,
            this.responseData.body[question.questionId]
          )
        )
      })
    },
    sendResponse(data) {
      // サーバーにPOST/PATCHリクエストを送る
      if (this.isNewResponse) {
        return axios
          .post('/responses', data)
          .then(resp => {
            const responseId = resp.data.responseID
            this.showMessage()
            router.push({
              name: 'ResponseDetails',
              params: { id: responseId }
            })
          })
          .catch(error => {
            console.log(error)
            this.alertNetworkError()
          })
      } else {
        return axios
          .patch('/responses/' + this.responseId, data)
          .then(this.getResponseData)
          .then(this.setResponsesToQuestions)
          .then(() => {
            this.showMessage()
            this.isEditing = false
          })
          .catch(error => {
            console.log(error)
            this.alertNetworkError()
          })
      }
    },
    submitResponse() {
      if (this.isSubmitting) return // 二重サブミット防止

      // 回答の送信
      let data = this.createResponseData()
      data.submitted_at = new Date().toLocaleString('ja-GB')

      this.isSubmitting = true
      this.sendResponse(data).then(() => {
        this.isSubmitting = false
        this.setMessage('回答を送信しました', 'green')
      })
    },
    saveResponse() {
      if (this.isSaving) return // 二重サブミット防止

      // 回答の保存
      let data = this.createResponseData()
      data.submitted_at = 'NULL'

      this.isSaving = true
      this.sendResponse(data).then(() => {
        this.isSaving = false
        this.setMessage('回答を保存しました (まだ未送信です)', 'green')
      })
    },
    abortEditing() {
      // TODO: 変更したかどうかを検出
      // const alertMessage = this.isNewResponse ? '回答を破棄します。よろしいですか？' : '変更を破棄します。よろしいですか？'
      // if (window.confirm(alertMessage)) {
      if (this.isNewResponse) {
        // 新しい回答の場合は、アンケートの詳細画面に戻る
        router.push({
          name: 'QuestionnaireDetails',
          params: { id: this.questionnaireId }
        })
      } else {
        this.isEditing = false
        this.getResponseData()
          .then(this.getQuestionnaireData)
          .then(this.getQuestions)
          .then(this.setResponsesToQuestions)
      }
      // }
    },
    createResponseData() {
      // サーバーに送るフォーマットのresponseDataを作成する
      let data = {
        questionnaireID: this.questionnaireId,
        body: []
      }
      this.questions.forEach(question => {
        let body = {
          questionID: question.questionId,
          question_type: question.type,
          response: '',
          option_response: []
        }
        switch (question.type) {
          case 'MultipleChoice':
            body.option_response = [question.selected]
            break
          case 'Checkbox':
            Object.keys(question.isSelected).forEach(key => {
              if (question.isSelected[key]) {
                body.option_response.push(key)
              }
            })
            break
          case 'Text':
          case 'Number':
            body.option_response = []
            body.response = String(question.responseBody)
            break
          case 'LinearScale':
            body.option_response = []
            body.response = String(question.selected)
            break
        }
        data.body.push(body)
      })
      return data
    },
    hasAnswered(question) {
      let hasSelectedOption = false
      switch (question.type) {
        case 'Text':
        case 'Number':
          return (
            typeof question.responseBody !== 'undefined' &&
            question.responseBody !== ''
          )
        case 'Checkbox':
          for (const option of Object.keys(question.isSelected)) {
            if (question.isSelected[option]) {
              hasSelectedOption = true
            }
          }
          return hasSelectedOption
        case 'MultipleChoice':
        case 'LinearScale':
          return (
            typeof question.selected !== 'undefined' && question.selected !== ''
          )
        default:
          return true
      }
    },
    setMessage(body, color) {
      this.$set(this.message, 'color', color)
      this.$set(this.message, 'body', body)
    },
    showMessage() {
      this.$set(this.message, 'showMessage', true)
    },
    resetMessage() {
      this.message = {
        showMessage: false
      }
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.details-child.is-fullheight {
  min-height: -webkit-fill-available;
}
.error-message {
  font-size: 1rem;
  margin: 1rem;
}
.tabs {
  ul {
    height: 2.5rem;
  }
}
</style>
