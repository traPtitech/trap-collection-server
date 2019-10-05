<template>
  <div class="columns is-flex">
    <article class="column is-11">
      <div class="card">
        <!-- タイトル、説明、回答期限 -->
        <div>
          <header class="card-header">
            <div id="title" class="card-header-title title">
              <span
                v-if="information.responseIconClass"
                :class="information.responseIconClass"
                class="response-icon"
              ></span>
              <router-link
                v-if="information.titleLink"
                :to="information.titleLink"
                >{{ information.title }}</router-link
              >
              <p v-if="!information.titleLink">{{ information.title }}</p>
            </div>
          </header>

          <div v-if="information.description" class="card-content">
            <pre>{{ information.description }}</pre>
          </div>

          <div
            v-if="information.timeLimit"
            class="is-pulled-right wrapper editable time-limit"
          >
            <span class="label">回答期限 :</span>
            <span class="time-limit">{{ information.timeLimit }}</span>
          </div>

          <div v-if="information.responseDetails" class="response-details">
            <div class="is-pulled-left is-inline-block wrapper editable">
              <span class="respondent subtitle is-5">{{
                information.responseDetails.respondent
              }}</span>
            </div>
            <div class="is-pulled-right is-flex editable">
              <span class="label">{{
                information.responseDetails.timeLabel + ': '
              }}</span>
              <span class="time-limit">{{
                information.responseDetails.time
              }}</span>
            </div>
          </div>
        </div>
      </div>
    </article>
  </div>
</template>

<script>
// import <componentname> from '<path to component file>'

export default {
  name: '',
  components: {},
  props: {
    information: {
      type: Object,
      required: true
    }
  },
  data() {
    return {}
  },
  computed: {},
  mounted() {},
  methods: {}
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.card {
  max-width: 100%;
  padding: 0.7rem;
}
.columns {
  margin: 0;
}
.column {
  padding: 0;
}
.editable {
  span {
    width: fit-content;
    height: fit-content;
    top: 0;
    bottom: 0;
    margin: auto 0.2rem;
    white-space: nowrap;
  }
  &.wrapper {
    display: flex;
  }
}
.response-icon {
  font-size: large;
  margin-right: 0.5rem;
}
.response-details,
.time-limit {
  margin: 1rem;
}
pre {
  white-space: pre-line;
  font-size: inherit;
  -webkit-font-smoothing: inherit;
  font-family: inherit;
  line-height: inherit;
  background-color: inherit;
  color: inherit;
  padding: 0.625em;
}
</style>
