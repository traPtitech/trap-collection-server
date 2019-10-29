<template>
  <div>
    <router-link
      v-if="type === 'newResponse' || type === 'viewResults'"
      class="button"
      :class="{ 'is-disabled': disabled }"
      :to="routeData"
    >
      <span :class="iconClass"></span>
      <span v-if="size === 'normal'">{{ buttonLabel }}</span>
    </router-link>
    <div
      v-if="type === 'deleteQuestionnaire'"
      :class="{ 'is-disabled': disabled || processing }"
      :disabled="disabled || processing"
      class="button"
      @click.prevent="deleteQuestionnaire"
    >
      <span :class="iconClass"></span>
      <span v-if="size === 'normal'">{{ buttonLabel }}</span>
    </div>
    <button
      v-if="type === 'closeQuestionnaire'"
      :class="{ 'is-disabled': disabled | processing }"
      class="button"
      @click.prevent="closeQuestionnaire"
    >
      <span :class="iconClass"></span>
      <span v-if="size === 'normal'">{{ buttonLabel }}</span>
    </button>
  </div>
</template>

<script>
import axios from '@/bin/axios'
import moment from 'moment'

export default {
  name: 'ManagementButton',
  components: {},
  props: {
    questionnaireId: {
      type: Number,
      default: undefined
    },
    size: {
      type: String,
      default: 'normal'
    },
    type: {
      type: String,
      default: 'newResponse'
    },
    disabled: {
      type: Boolean,
      default: false
    },
    questionnaireInformation: {
      type: Object,
      default: undefined
    }
  },
  data() {
    return {
      processing: false,
      iconClasses: {
        newResponse: 'ti-check-box',
        viewResults: 'ti-bar-chart',
        closeQuestionnaire: 'ti-timer',
        deleteQuestionnaire: 'ti-trash'
      },
      buttonLabels: {
        newResponse: '回答する',
        viewResults: '結果を見る',
        closeQuestionnaire: 'アンケートを締め切る',
        deleteQuestionnaire: 'アンケートを削除'
      }
    }
  },
  computed: {
    iconClass() {
      return this.iconClasses[this.type]
    },
    buttonLabel() {
      return this.buttonLabels[this.type]
    },
    routeData() {
      switch (this.type) {
        case 'newResponse':
          return {
            name: 'NewResponseDetails',
            params: { questionnaireId: this.questionnaireId }
          }
        case 'viewResults':
          return {
            name: 'Results',
            params: { id: this.questionnaireId }
          }
        default:
          console.error('no Route Data')
          return null
      }
    }
  },
  watch: {},
  mounted() {},
  methods: {
    deleteQuestionnaire() {
      if (this.disabled || this.processing) return
      if (window.confirm('アンケートを削除しますか？')) {
        if (this.isNewQuestionnaire) {
          this.$router.push('/administrates')
        } else {
          this.processing = true
          axios
            .delete('/api/questionnaires/' + this.questionnaireId)
            .then(() => {
              this.processing = false
              this.$router.push('/administrates')
              // アンケートを削除したら、Administratesページに戻る
            })
            .catch(error => {
              this.processing = false
              console.log(error)
              this.alertNetworkError()
            })
            .finally(() => {})
        }
      }
    },

    closeQuestionnaire() {
      if (this.disabled || this.processing) return
      if (window.confirm('アンケートを今すぐ締め切りますか？')) {
        this.processing = true
        const time_limit = moment().format('YYYY/MM/DD HH:mm')
        axios
          .patch('/api/questionnaires/' + this.questionnaireId, {
            title: this.questionnaireInformation.title,
            description: this.questionnaireInformation.description,
            res_time_limit: time_limit,
            res_shared_to: this.questionnaireInformation.res_shared_to,
            targets: this.questionnaireInformation.targets,
            administrators: this.questionnaireInformation.administrators
          })
          .then(() => {
            this.processing = false
            this.questionnaireInformation.res_time_limit = time_limit
          })
          .catch(error => {
            this.processing = false
            console.log(error)
            this.alertNetworkError()
          })
          .finally(() => {})
      }
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.button {
  max-width: fit-content;
  display: block;
  &:hover {
    background-color: $base-pink;
  }
}
</style>
