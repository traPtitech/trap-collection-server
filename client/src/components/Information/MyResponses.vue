<template>
  <div>
    <header class="card-header">
      <div class="card-header-title subtitle">自分の回答</div>
    </header>
    <div class="card-content">
      <ul class="response-list">
        <li
          v-for="(response, index) in responses"
          :key="index"
          class="response-list-item"
        >
          <span
            :class="{
              'ti-save': response.submitted_at === 'NULL',
              'ti-check': response.submitted_at !== 'NULL'
            }"
          ></span>
          <router-link
            :to="{
              name: 'ResponseDetails',
              params: { id: response.responseID }
            }"
            >{{ getDateStr(response.modified_at) }}</router-link
          >
          <a>
            <span
              class="ti-trash is-pulled-right"
              @click="deleteResponse(response.responseID, index)"
            ></span>
          </a>
        </li>
      </ul>
    </div></div
></template>

<script>
import axios from '@/bin/axios'
import common from '@/bin/common'

export default {
  name: 'MyResponses',
  components: {},
  props: {
    questionnaireId: {
      type: Number,
      default: undefined
    }
  },
  data() {
    return {
      responses: [],
      processing: {}
    }
  },
  computed: {},
  watch: {
    responses: function(newArr) {
      // 回答を送信済みかどうかを調べて Information に送信
      let hasResponded = false
      newArr.forEach(response => {
        if (response.submitted_at !== 'NULL') hasResponded = true
      })
      this.$emit('set-has-responded', hasResponded)
    }
  },
  mounted() {},
  methods: {
    getDateStr: common.getDateStr,
    async deleteResponse(responseId, index) {
      if (this.processing[responseId]) return
      if (window.confirm('この回答を削除しますか？')) {
        this.processing[responseId] = true
        await axios
          .delete('/api/responses/' + responseId, {
            method: 'delete',
            withCredentials: true
          })
          .then(() => {
            this.responses.splice(index, 1)
            this.processing[responseId] = false
          })
      }
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.response-list-item:hover {
  background: $base-gray;
}
</style>
