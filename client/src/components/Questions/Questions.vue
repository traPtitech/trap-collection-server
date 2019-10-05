<template>
  <div class="columns">
    <article class="column is-11">
      <div class="card">
        <div class="card-content questions">
          <div
            v-for="(question, index) in questions"
            :key="index"
            class="question"
          >
            <div class="question-body">
              <p class="subtitle">
                {{ question.questionBody }}
                <span
                  v-if="showRequiredIcon(index)"
                  class="ti-alert required-question-icon"
                  >必須</span
                >
              </p>
            </div>
            <input-error-message
              v-if="inputErrors"
              :input-error="inputErrors[question.questionId]"
            ></input-error-message>
            <component
              :is="question.component"
              :edit-mode="editMode"
              :content-props="question"
              :question-index="index"
              class="response-body"
              @set-question-content="setQuestionContent"
            ></component>
            <hr />
          </div>
        </div>
      </div>
    </article>
  </div>
</template>

<script>
import MultipleChoice from '@/components/Questions/MultipleChoice'
import LinearScale from '@/components/Questions/LinearScale'
import ShortAnswer from '@/components/Questions/ShortAnswer'
import InputErrorMessage from '@/components/Utils/InputErrorMessage'

// import common from '@/bin/common'

export default {
  name: 'Questions',
  components: {
    'multiple-choice': MultipleChoice,
    'linear-scale': LinearScale,
    'short-answer': ShortAnswer,
    'input-error-message': InputErrorMessage
  },
  props: {
    questionsProps: {
      type: Array,
      required: false,
      default: undefined
    },
    editMode: {
      type: String,
      required: false,
      default: undefined
    },
    inputErrors: {
      type: Object,
      required: false,
      default: undefined
    }
  },
  data() {
    return {}
  },
  computed: {
    questions() {
      return this.questionsProps
    }
  },
  mounted() {},
  methods: {
    showRequiredIcon(index) {
      return this.questions[index].isRequired
    },
    setQuestionContent(index, label, value) {
      this.$emit('set-question-content', index, label, value)
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.required-question-icon {
  font-size: 0.8rem;
  color: red;
  margin: auto 0.5rem;
  word-break: keep-all;
}
.required-question-icon::before {
  margin-right: 0.2rem;
}
</style>
