<template>
  <div>
    <!-- view only -->
    <div v-if="!editMode">
      <div
        v-for="(option, index) in content.options"
        :key="index"
        class="is-flex"
      >
        <span
          :class="[
            readOnlyBoxClass,
            isSelected(option.label) ? 'checked' : 'empty'
          ]"
        ></span>
        <span class="option-label">{{ option.label }}</span>
      </div>
    </div>

    <!-- edit question -->
    <div v-if="editMode === 'question'">
      <transition-group name="list" tag="div">
        <div
          v-for="(option, index) in content.options"
          :key="option.id"
          class="is-flex"
        >
          <span class="sort-handle">
            <span
              :class="{ disabled: isFirstOption(index) }"
              class="ti-angle-up icon"
              @click="swapOrder(content.options, index, index - 1)"
            ></span>
            <span
              :class="{ disabled: isLastOption(index) }"
              class="ti-angle-down icon"
              @click="swapOrder(content.options, index, index + 1)"
            ></span>
          </span>
          <span :class="readOnlyBoxClass"></span>
          <input
            :value="content.options[index].label"
            type="text"
            class="input has-underline option-label is-editable"
            @input="setOption(index, $event.target.value)"
          />
          <span class="delete-button">
            <span
              class="ti-trash icon is-medium"
              @click="removeOption(index)"
            ></span>
          </span>
        </div>
      </transition-group>
      <div class="wrapper add-option">
        <div class="add-option-button" @click="addOption()">
          <span class="ti-plus circled icon"></span>
          <span>新しい選択肢を追加</span>
        </div>
      </div>
    </div>

    <!-- edit response -->
    <div v-if="editMode === 'response'">
      <div
        v-for="(option, index) in content.options"
        :key="index"
        class="is-flex"
      >
        <label class="option-label">
          <input
            v-if="content.type === 'Checkbox'"
            v-model="contentProps.isSelected[option.label]"
            :value="option.label"
            type="checkbox"
          />
          <input
            v-if="content.type === 'MultipleChoice'"
            v-model="contentProps.selected"
            :value="option.label"
            type="radio"
          />
          {{ option.label }}
        </label>
      </div>
    </div>
  </div>
</template>

<script>
// import <componentname> from '<path to component file>'
import common from '@/bin/common'
import question from '@/bin/question'

export default {
  name: 'MultipleChoice',
  components: {},
  props: {
    contentProps: {
      type: Object,
      required: true
    },
    questionIndex: {
      type: Number,
      required: false,
      default: undefined
    },
    editMode: {
      type: String,
      required: false,
      default: undefined
    }
  },
  data() {
    return {
      newId: -1
    }
  },
  computed: {
    content() {
      return this.contentProps
    },
    readOnlyBoxClass() {
      switch (this.content.type) {
        case 'Checkbox':
          return 'readonly-checkbox'
        case 'MultipleChoice':
          return 'readonly-radiobutton'
        default:
          return undefined
      }
    }
  },
  watch: {},
  mounted() {},
  methods: {
    swapOrder: common.swapOrder,
    setContent: question.setContent,
    setOption(index, value) {
      let newOptions = Object.assign([], this.content.options)
      newOptions[index].label = value
      this.setContent('options', newOptions)
    },
    isFirstOption(index) {
      return index === 0
    },
    isLastOption(index) {
      return index === this.content.options.length - 1
    },
    isSelected(label) {
      if (this.content.type === 'Checkbox') {
        return this.content.isSelected[label]
      } else if (this.content.type === 'MultipleChoice') {
        return this.content.selected === label
      }
    },
    updateOptionId() {
      this.newId--
    },
    addOption() {
      let newOptions = Object.assign([], this.content.options)
      newOptions.push({
        id: this.newId,
        label: ''
      })
      this.updateOptionId()
      this.setContent('options', newOptions)
    },
    removeOption(index) {
      let newOptions = Object.assign([], this.content.options)
      newOptions.splice(index, 1)
      this.setContent('options', newOptions)
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
input[type='checkbox'] {
  margin: auto 0;
}
.option-label {
  width: 100%;
  padding: 0 0.5rem;
  margin: auto;
}
.sort-handle {
  margin-right: 0.7rem;
}
.delete-button {
  height: fit-content;
  margin: auto;
}
.wrapper.add-option {
  display: flex;
  height: 2.5rem;
}
.add-option-button {
  margin: auto;
  cursor: pointer;
}
</style>
