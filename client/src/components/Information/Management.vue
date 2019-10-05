<template>
  <div>
    <header class="card-header">
      <div class="card-header-title subtitle">操作</div>
    </header>
    <div class="card-content management-buttons">
      <!-- 回答する -->
      <management-button
        :disabled="timeLimitExceeded"
        :questionnaire-id="questionnaireId"
        type="newResponse"
        class="button-wrapper"
      >
      </management-button>
      <div class="new-response-link-panel">
        <input
          id="new-response-link"
          ref="link"
          :value="newResponseLink"
          class="input"
          type="text"
          readonly
          @click="$refs.link.select()"
        />
        <span class="button" @click="copyNewResponseLink">
          <span class="ti-clipboard"></span>
        </span>
      </div>
      <transition name="fade">
        <p v-if="copyMessage.showMessage" class="copy-message">
          {{ copyMessage.message }}
        </p>
      </transition>

      <!-- 結果を見る -->
      <management-button
        :disabled="!canViewResults"
        :questionnaire-id="questionnaireId"
        class="button-wrapper"
        type="viewResults"
      >
      </management-button>

      <!-- アンケートを締め切る -->
      <management-button
        :disabled="!administrates || timeLimitExceeded"
        :questionnaire-id="questionnaireId"
        :questionnaire-information="questionnaireInformation"
        class="button-wrapper"
        type="closeQuestionnaire"
      >
      </management-button>

      <!-- アンケートを削除 -->
      <management-button
        :disabled="!administrates"
        :questionnaire-id="questionnaireId"
        class="button-wrapper"
        type="deleteQuestionnaire"
      >
      </management-button>
    </div>
  </div>
</template>

<script>
import ManagementButton from '@/components/Information/ManagementButton'

export default {
  name: 'Management',
  components: {
    'management-button': ManagementButton
  },
  props: {
    questionnaireInformation: {
      type: Object,
      required: true
    },
    questionnaireId: {
      type: Number,
      default: undefined
    },
    canViewResults: {
      type: Boolean
    },
    administrates: {
      type: Boolean
    }
  },
  data() {
    return {
      copyMessage: {
        showMessage: false
      }
    }
  },
  computed: {
    timeLimitExceeded() {
      // 回答期限を過ぎていた場合はtrueを返す
      return (
        new Date(this.questionnaireInformation.res_time_limit).getTime() <
        new Date().getTime()
      )
    },
    newResponseLink() {
      return (
        location.protocol +
        '//' +
        location.host +
        '/responses/new/' +
        this.questionnaireId
      )
    }
  },
  watch: {},
  mounted() {},
  methods: {
    copyNewResponseLink() {
      let link = document.querySelector('#new-response-link')
      // link.select()
      let range = document.createRange()
      range.selectNode(link)
      const selection = window.getSelection()
      selection.removeAllRanges()
      selection.addRange(range)
      if (document.execCommand('copy')) {
        this.showCopyMessage('リンクをコピーしました！')
      } else {
        this.showCopyMessage('コピーに失敗しました')
      }
    },
    async showCopyMessage(message) {
      this.copyMessage = {
        showMessage: true,
        message: message
      }
      await new Promise(resolve => setTimeout(resolve, 3000))
      this.resetCopyMessage()
    },
    resetCopyMessage() {
      this.copyMessage = {
        showMessage: false
      }
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.management-buttons {
  .new-response-link-panel {
    display: flex;
    input {
      max-width: 20rem;
      margin: 0;
      display: inherit;
      border-radius: 4px 0 0 4px;
      border-color: $base-gray;
    }
    .button {
      min-width: fit-content;
      margin: 0;
      display: inline-block;
      border-radius: 0 4px 4px 0;
      border-color: $base-gray;
      &:hover {
        background-color: $base-pink;
      }
    }
  }
  .copy-message {
    font-size: smaller;
    margin: 0.3rem;
  }
  .button-wrapper:not(:first-child) {
    margin-top: 0.7rem;
  }
}
</style>
