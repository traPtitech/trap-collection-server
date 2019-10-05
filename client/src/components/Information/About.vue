<template>
  <div>
    <header class="card-header">
      <div class="card-header-title subtitle">情報</div>
    </header>
    <div class="card-content">
      <div class="is-flex res-shared-to">
        <span class="label">結果:</span>
        <span>{{ resSharedToLabel }}</span>
      </div>

      <!-- user lists -->
      <div class="user-lists">
        <div v-for="(userList, key) in userLists" :key="key">
          <div
            class="label"
            @click.prevent="toggleListVisibility(userList.name)"
          >
            <span
              :class="userList.show ? 'ti-angle-down' : 'ti-angle-right'"
            ></span>
            <span>{{ userList.summary }} ({{ userList.list.length }})</span>
          </div>
          <p v-if="userList.show" class="has-text-grey user-list">
            <span v-for="(user, index) in userList.list" :key="index">
              <span
                :class="{
                  'highlight-name': user === 'traP' || user === getMyTraqId
                }"
                >{{ user }}</span
              >
              <span>{{ index === userList.list.length - 1 ? '' : ', ' }}</span>
            </span>
          </p>
        </div>
      </div>

      <div class="has-text-weight-bold">
        <div>更新日時 : {{ getDateStr(modifiedAt) }}</div>
        <div>作成日時 : {{ getDateStr(createdAt) }}</div>
      </div>
    </div>
  </div>
</template>

<script>
/* eslint-disable vue/require-default-prop */
// TODO: administrators, respondents, targets の管理をstoreで行うようにする

import common from '@/bin/common'
import { mapGetters } from 'vuex'

export default {
  name: 'About',
  components: {},
  props: {
    resSharedTo: {
      type: String,
      default: undefined
    },
    administrators: {
      type: Object
      // validator: this.listValidator,
      // default: this.defaultListObj
    },
    respondents: {
      type: Object
      // validator: this.listValidator,
      // default: this.defaultListObj
    },
    targets: {
      type: Object
      // validator: this.listValidator,
      // default: this.defaultListObj
    },
    modifiedAt: {
      type: String,
      default: undefined
    },
    createdAt: {
      type: String,
      default: undefined
    }
  },
  data() {
    return {
      defaultListObj: {
        list: [],
        editable: false,
        name: '',
        show: false,
        summary: ''
      }
    }
  },
  computed: {
    ...mapGetters(['getMyTraqId']),
    resSharedToLabel() {
      const labels = {
        public: '全体に公開',
        respondents: '回答済みの人に公開',
        administrators: '管理者のみに公開'
      }
      return labels[this.resSharedTo]
    },
    userLists() {
      return {
        administrators: this.administrators
          ? this.administrators
          : this.defaultListObj,
        respondents: this.respondents ? this.respondents : this.defaultListObj,
        targets: this.targets ? this.targets : this.defaultListObj
      }
    }
  },
  watch: {},
  methods: {
    getDateStr: common.getDateStr,
    listValidator(listObj) {
      if (typeof listObj === 'undefined') return true
      if (!Array.isArray(listObj.list)) return false
      listObj.list.forEach(item => {
        if (typeof item !== 'string') return false
      })

      return (
        typeof listObj.editable === 'boolean' &&
        typeof listObj.name === 'string' &&
        typeof listObj.show === 'boolean' &&
        typeof listObj.summary === 'string'
      )
    },
    toggleListVisibility(listName) {
      this.userLists[listName].show = !this.userLists[listName].show
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.res-shared-to {
  span {
    width: fit-content;
    height: fit-content;
    margin: auto 0.2rem;
  }
}
.user-lists {
  margin: 0.5rem;
  .label {
    margin: 0.5rem 0 0.1rem 0;
    display: flex;
    span {
      margin: auto 0.1rem;
    }
    cursor: pointer;
  }
}
</style>
