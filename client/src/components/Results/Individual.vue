<template>
  <div class="has-navbar-fixed-bottom">
    <questions :questions-props="questionData"></questions>
    <pagination
      :range="range"
      :current-page="currentPage"
      :default-page-link="defaultPageLink"
    ></pagination>
  </div>
</template>

<script>
import Questions from '@/components/Questions/Questions'
import Pagination from '@/components/Utils/Pagination'

export default {
  name: 'Individual',
  components: {
    questions: Questions,
    pagination: Pagination
  },
  props: {
    results: {
      type: Array,
      required: true
    },
    responseData: {
      type: Object,
      required: true
    },
    questionData: {
      type: Array,
      required: true
    },
    information: {
      type: Object,
      required: true
    }
  },
  data() {
    return {}
  },
  computed: {
    currentPage() {
      return this.$route.query.page
        ? Number(this.$route.query.page)
        : this.range.first
    },
    range() {
      return {
        first: 1,
        last: this.results.length
      }
    },
    defaultPageLink() {
      return {
        name: 'Results',
        params: { id: this.$route.params.id },
        query: { tab: 'individual' }
      }
    }
  },
  watch: {},
  async created() {},
  methods: {}
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped></style>
