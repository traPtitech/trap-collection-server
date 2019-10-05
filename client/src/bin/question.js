/* eslint-disable */
// LinearScale, MultipleChoice, ShortAnswer から呼び出す
export default {
  setContent: function(label, value) {
    this.$emit('set-question-content', this.questionIndex, label, value)
  }
}
