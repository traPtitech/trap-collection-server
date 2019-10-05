<template>
  <div>
    <!-- view only -->
    <div v-if="!editMode">
      <div class="columns">
        <div class="column">{{ content.scaleLabels.left }}</div>
        <div class="column is-9 is-9-mobile is-flex">
          <span
            v-for="(num, index) in scaleArray"
            :key="index"
            class="scale-num has-text-centered"
          >
            <div>{{ num }}</div>
            <div>
              <span
                class="readonly-radiobutton"
                :class="{ checked: isSelected(num) }"
              ></span>
            </div>
          </span>
        </div>
        <div class="column has-text-right">{{ content.scaleLabels.right }}</div>
      </div>
    </div>

    <!-- edit question -->
    <div v-if="editMode === 'question'">
      <div class="is-flex scale-range-edit">
        <span class="select">
          <select
            :value="content.scaleRange.left"
            @input="setScaleRange('left', Number($event.target.value))"
          >
            <option
              v-for="(num, index) in progressiveArray(0, 1)"
              :key="index"
              >{{ num }}</option
            >
          </select>
        </span>
        <span>to</span>
        <span class="select">
          <select
            :value="content.scaleRange.right"
            @input="setScaleRange('right', Number($event.target.value))"
          >
            <option
              v-for="(num, index) in progressiveArray(2, 10)"
              :key="index"
              >{{ num }}</option
            >
          </select>
        </span>
      </div>
      <div class="scale-label-edit is-flex">
        <span>{{ content.scaleRange.left }}</span>
        <span>
          <input
            :value="content.scaleLabels.left"
            type="text"
            placeholder="ラベル (任意)"
            class="input has-underline is-editable"
            @input="setScaleLabels('left', $event.target.value)"
          />
        </span>
      </div>
      <div class="scale-label-edit is-flex">
        <span>{{ content.scaleRange.right }}</span>
        <span>
          <input
            :value="content.scaleLabels.right"
            type="text"
            placeholder="ラベル (任意)"
            class="input has-underline is-editable"
            @input="setScaleLabels('right', $event.target.value)"
          />
        </span>
      </div>
    </div>

    <!-- edit response -->
    <div v-if="editMode === 'response'">
      <div class="columns">
        <div class="column">{{ content.scaleLabels.left }}</div>
        <div class="column is-9 is-9-mobile is-flex">
          <span
            v-for="(num, index) in scaleArray"
            :key="index"
            class="scale-num has-text-centered"
          >
            <label>
              {{ num }}
              <input
                v-model="contentProps.selected"
                :value="num"
                type="radio"
              />
            </label>
          </span>
        </div>
        <div class="column has-text-right">{{ content.scaleLabels.right }}</div>
      </div>
    </div>
  </div>
</template>

<script>
// import <componentname> from '<path to component file>'
import question from '@/bin/question.js'

export default {
  name: 'LinearScale',
  components: {},
  props: {
    questionIndex: {
      type: Number,
      required: false,
      default: undefined
    },
    contentProps: {
      type: Object,
      required: true
    },
    editMode: {
      type: String,
      required: false,
      default: undefined
    }
  },
  data() {
    return {}
  },
  computed: {
    content() {
      return this.contentProps
    },
    scaleArray() {
      return this.progressiveArray(
        this.content.scaleRange.left,
        this.content.scaleRange.right
      )
    }
  },
  mounted() {},
  methods: {
    setContent: question.setContent,
    setScaleRange(side, value) {
      let newScaleRange = Object.assign({}, this.content.scaleRange)
      newScaleRange[side] = value
      this.setContent('scaleRange', newScaleRange)
    },
    setScaleLabels(side, value) {
      let newScaleLabels = Object.assign({}, this.content.scaleLabels)
      newScaleLabels[side] = value
      this.setContent('scaleLabels', newScaleLabels)
    },
    isSelected(num) {
      return this.content.selected === num
    },
    progressiveArray(min, max) {
      const len = max - min + 1
      let arr = []
      for (let i = 0; i < len; i++) {
        arr[i] = min + i
      }
      return arr
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.column {
  padding: 0.5rem;
}
.scale-num {
  margin: auto;
  width: min-content;
}
.scale-range-edit,
.scale-label-edit {
  span {
    margin: auto 0.5rem;
  }
}
.scale-label-edit {
  margin: 1rem 0.5rem;
}
</style>
