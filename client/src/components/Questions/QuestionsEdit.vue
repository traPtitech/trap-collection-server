<template>
  <div class="columns is-editing">
    <article class="column is-11">
      <div class="card">
        <input-error-message
          :input-error="inputErrors.noQuestions"
        ></input-error-message>
        <transition-group name="list" tag="div" class="card-content questions">
          <div
            v-for="(question, index) in questions"
            :key="question.questionId"
            class="question"
          >
            <div class="question-body columns is-mobile">
              <div class="column left-bar">
                <div class="sort-handle">
                  <span
                    class="ti-angle-up icon"
                    :class="{ disabled: isFirstQuestion(index) }"
                    @click="swapOrder(questions, index, index - 1)"
                  ></span>
                  <span
                    class="ti-angle-down icon"
                    :class="{ disabled: isLastQuestion(index) }"
                    @click="swapOrder(questions, index, index + 1)"
                  ></span>
                </div>
                <div class="delete-button">
                  <span
                    class="ti-trash icon is-medium"
                    @click="removeQuestion(index)"
                  ></span>
                </div>
              </div>
              <div class="column question is-editable">
                <div class="columns is-inline-block-mobile">
                  <div class="column">
                    <input
                      v-model="question.questionBody"
                      type="text"
                      class="subtitle input has-underline is-editable"
                      placeholder="質問文"
                    />
                  </div>
                  <div
                    class="column is-2 required-question-checkbox is-pulled-right"
                  >
                    <label class="checkbox">
                      必須
                      <input v-model="question.isRequired" type="checkbox" />
                    </label>
                  </div>
                </div>
                <component
                  :is="question.component"
                  :edit-mode="'question'"
                  :content-props="question"
                  :question-index="index"
                  class="response-body"
                  @set-question-content="setQuestionContent"
                ></component>
              </div>
            </div>
            <hr />
          </div>
        </transition-group>
        <div class="add-question">
          <div
            class="add-question-button button"
            @click="toggleNewQuestionDropdown"
          >
            <span>新しい質問を追加</span>
            <span
              class="icon is-small"
              :class="
                newQuestionDropdownIsActive ? 'ti-angle-down' : 'ti-angle-right'
              "
            ></span>
          </div>
          <div
            v-show="newQuestionDropdownIsActive"
            class="question-type-buttons"
          >
            <button
              v-for="(questionType, key) in questionTypes"
              :key="key"
              class="button"
              @click="insertQuestion(questionType)"
            >
              {{ questionType.label }}
            </button>
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
import common from '@/bin/common'

export default {
  name: 'QuestionsEdit',
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
    title: {
      type: String,
      required: true
    },
    inputErrors: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      newQuestionDropdownIsActive: false,
      questionTypes: common.questionTypes,
      lastQuestionId: 0
    }
  },
  computed: {
    questions() {
      return this.questionsProps
    }
  },
  mounted() {},
  methods: {
    swapOrder: common.swapOrder,
    setQuestions(questions) {
      this.$emit('set-data', 'questions', questions)
    },
    setQuestionContent(index, label, value) {
      this.$emit('set-question-content', index, label, value)
    },
    isFirstQuestion(index) {
      return index === 0
    },
    isLastQuestion(index) {
      return index === this.questions.length - 1
    },
    removeQuestion(index) {
      this.$emit('remove-question', index)
    },
    insertQuestion(questionType) {
      this.questions.push(this.getDefaultQuestion(questionType))
      this.setQuestions(this.questions)
    },
    toggleNewQuestionDropdown() {
      this.newQuestionDropdownIsActive = !this.newQuestionDropdownIsActive
    },
    getDefaultQuestion(questionType) {
      let ret = {
        questionId: this.getNewQuestionId(),
        type: questionType.type,
        component: questionType.component,
        questionBody: '',
        isRequired: false,
        pageNum: 1 // ぺージ分けは未実装
      }
      switch (questionType.type) {
        case 'Checkbox':
          ret.options = [{ label: '', id: 0 }]
          ret.isSelected = [false]
          break
        case 'MultipleChoice':
          ret.options = [{ label: '', id: 0 }]
          break
        case 'LinearScale':
          ret.scaleRange = { left: 1, right: 5 }
          ret.scaleLabels = { left: '', right: '' }
          break
      }
      return ret
    },
    getNewQuestionId() {
      const ret = this.lastQuestionId - 1
      this.lastQuestionId -= 1
      return ret
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.checkbox {
  // 1行に1つの選択肢
  display: block;
}
.card-content {
  padding: 1rem 0 0 0;
}
.question-body {
  .column {
    padding: 0.5rem;
  }
}
.question.is-editable {
  border-left: solid;
  padding-left: 1rem;
  padding: 0 0.5rem;
}
.wrapper.add-option {
  display: flex;
  height: 2.5rem;
}
.add-option-button {
  margin: auto;
  cursor: pointer;
}
.required-question-checkbox {
  margin: auto;
  label.checkbox {
    padding: 0.5rem;
    // margin: 0 0.2rem;
    width: fit-content;
    border-left: solid 1px gray;
  }
}
.sort-handle {
  height: 5rem;
}
.add-question {
  margin-bottom: 1rem;
}
.add-question-button {
  background: whitesmoke;
  .icon {
    margin-right: 0.3rem;
    margin-left: 0.3rem;
  }
}
.question-type-buttons {
  background: whitesmoke;
  padding: 0.5rem;
  border-radius: 0.2rem;
  button {
    margin: 0.2rem;
  }
}
.question:first-child {
  padding-top: 1rem;
}
.details-child.is-fullheight {
  min-height: -webkit-fill-available;
}
</style>
