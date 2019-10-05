<template>
  <nav class="navbar is-fixed-bottom pagination is-rounded is-centered">
    <ul class="pagination-list">
      <li v-if="range">
        <router-link
          class="pagination-previous"
          :disabled="disableFirstButton"
          :to="getPageLink('first')"
        >
          <span class="ti-angle-double-left"></span>
        </router-link>
      </li>
      <li v-for="pageNum in pages" :key="pageNum">
        <router-link
          v-if="pageNum"
          class="pagination-link"
          :class="{ 'is-current': pageNum === currentPage }"
          :to="getPageLink(pageNum)"
          >{{ pageNum }}</router-link
        >
        <span v-if="!pageNum" class="pagination-link" disabled></span>
      </li>
      <li v-if="range">
        <router-link
          class="pagination-next"
          :disabled="disableLastButton"
          :to="getPageLink('last')"
        >
          <span class="ti-angle-double-right"></span>
        </router-link>
      </li>
    </ul>
  </nav>
</template>

<script>
// import <componentname> from '<path to component file>'

export default {
  name: 'Pagination',
  components: {},
  props: {
    range: {
      type: Object,
      required: false,
      default: undefined
    },
    currentPage: {
      type: Number,
      required: true
    },
    defaultPageLink: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      paginationWidth: 1
    }
  },
  computed: {
    pages() {
      let ret = []
      for (
        let i = this.currentPage - this.paginationWidth;
        i <= this.currentPage + this.paginationWidth;
        i++
      ) {
        let min = 1
        let max = this.currentPage + this.paginationWidth
        if (this.range) {
          min = this.range.first
          max = this.range.last
        }
        if (i >= min && i <= max) {
          ret.push(i)
        } else {
          ret.push(undefined)
        }
      }
      return ret
    },
    disableFirstButton() {
      return this.currentPage - this.paginationWidth <= this.range.first
    },
    disableLastButton() {
      return this.currentPage + this.paginationWidth >= this.range.last
    }
  },
  mounted() {},
  methods: {
    getPageLink(pageName) {
      let ret = Object.assign({}, this.defaultPageLink)
      ret.query =
        typeof this.defaultPageLink.query === 'undefined'
          ? {}
          : Object.assign({}, this.defaultPageLink.query)
      if (this.range && pageName === 'first') {
        ret.query.page = this.range.first
      } else if (this.range && pageName === 'last') {
        ret.query.page = this.range.last
      } else {
        ret.query.page = pageName
      }
      return ret
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.pagination {
  // margin: 0.5rem 0 0 0;
  padding: 0.75rem 0;
}
.pagination-next,
.pagination-previous {
  margin: 0.25rem 0.75rem;
}
.pagination-link[disabled] {
  cursor: default;
}
</style>
