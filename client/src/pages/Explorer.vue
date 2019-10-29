<template>
  <div class="wrapper is-fullheight has-navbar-fixed-bottom">
    <div class="dropdowns">
      <div
        class="dropdown"
        :class="{ 'is-active': DropdownIsActive.sortOrder }"
      >
        <div class="dropdown-trigger">
          <button
            class="button"
            aria-haspopup="true"
            aria-controls="dropdown-menu"
            @click="
              DropdownIsActive.targetedOption = false
              DropdownIsActive.sortOrder = !DropdownIsActive.sortOrder
            "
          >
            <p>並び替え</p>
            <span class="ti-angle-down"></span>
          </button>
        </div>
        <div id="dropdown-menu" class="dropdown-menu" role="menu">
          <div
            v-for="(order, index) in sortOrders"
            :key="index"
            :class="{ 'is-selected': order.opt === sortOrder }"
            class="dropdown-content"
            @click="
              changeSortOrder(order.opt)
              DropdownIsActive.sortOrder = false
            "
          >
            <p class="dropdown-item">{{ order.str }}</p>
          </div>
        </div>
      </div>
      <div
        class="dropdown"
        :class="{ 'is-active': DropdownIsActive.targetedOption }"
      >
        <div class="dropdown-trigger">
          <button
            class="button"
            aria-haspopup="true"
            aria-controls="dropdown-menu"
            @click="
              DropdownIsActive.sortOrder = false
              DropdownIsActive.targetedOption = !DropdownIsActive.targetedOption
            "
          >
            <p>フィルター</p>
            <span class="ti-angle-down"></span>
          </button>
        </div>
        <div id="dropdown-menu" class="dropdown-menu" role="menu">
          <div
            v-for="(option, index) in targetedOptions"
            :key="index"
            :class="{ 'is-selected': option.opt === targetedOption }"
            class="dropdown-content"
            @click="
              changetargetedOption(option.opt)
              DropdownIsActive.targetedOption = false
            "
          >
            <p class="dropdown-item">{{ option.str }}</p>
          </div>
        </div>
      </div>
    </div>
    <div
      class="card-wrapper is-fullheight"
      @click="
        DropdownIsActive.sortOrder = false
        DropdownIsActive.targetedOption = false
      "
    >
      <div class="card">
        <!-- <button class="button" v-on:click="changeSortOrder('-title')">Button</button> -->
        <table class="table is-striped">
          <thead>
            <tr>
              <th v-for="(header, index) in headers" :key="index">
                {{ header }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(questionnaire, index) in questionnaires" :key="index">
              <td class="table-item-title">
                <router-link
                  :to="'/questionnaires/' + questionnaire.questionnaireID"
                  >{{ questionnaire.title }}</router-link
                >
              </td>
              <td class="table-item-date">
                {{ getDateStr(questionnaire.res_time_limit) }}
              </td>
              <td class="table-item-date">
                {{ getRelativeDateStr(questionnaire.modified_at) }}
              </td>
              <td class="table-item-date">
                {{ getRelativeDateStr(questionnaire.created_at) }}
              </td>
              <td>
                <router-link
                  :to="'/results/' + questionnaire.questionnaireID"
                  target="_blank"
                >
                  <span class="ti-new-window"></span>
                  <br />Open
                </router-link>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <pagination
      :current-page="pageNumber"
      :default-page-link="defaultPageLink"
      :range="range"
    ></pagination>
  </div>
</template>

<script>
import axios from '@/bin/axios'
import router from '@/router'
import Pagination from '@/components/Utils/Pagination'
import common from '@/bin/common'

export default {
  name: 'Explorer',
  components: {
    pagination: Pagination
  },
  props: {},
  data() {
    return {
      questionnaires: [],
      headers: ['', '回答期限', '更新日時', '作成日時', '結果'],
      sortOrders: [
        {
          str: '最近更新された',
          opt: '-modified_at'
        },
        {
          str: '最近更新されていない',
          opt: 'modified_at'
        },
        {
          str: 'タイトル順',
          opt: 'title'
        },
        {
          str: 'タイトル逆順',
          opt: '-title'
        },
        {
          str: '最新',
          opt: '-created_at'
        },
        {
          str: '最も古い',
          opt: 'created_at'
        }
      ],
      targetedOptions: [
        {
          str: '全て',
          opt: false
        },
        {
          str: '対象外のもののみ',
          opt: true
        }
      ],
      DropdownIsActive: {
        sortOrder: false,
        targetedOption: false
      },
      range: {
        first: 1,
        last: 1
      }
    }
  },
  computed: {
    pageNumber: {
      get() {
        return this.$route.query.page ? Number(this.$route.query.page) : 1
      },
      set(newVal) {
        router.push({
          name: 'Explorer',
          query: {
            nontargeted: String(this.targetedOption),
            page: String(newVal),
            sort: this.sortOrder
          }
        })
      }
    },
    sortOrder: {
      get() {
        return this.$route.query.sort ? this.$route.query.sort : '-modified_at'
      },
      set(newVal) {
        router.push({
          name: 'Explorer',
          query: {
            nontargeted: String(this.targetedOption),
            page: String(this.pageNumber),
            sort: newVal
          }
        })
      }
    },
    targetedOption: {
      get() {
        // return typeof this.$route.query.nontargeted !== 'undefined' && this.$route.query.nontargeted === 'true'
        return this.$route.query.nontargeted === 'true'
      },
      set(newVal) {
        router.push({
          name: 'Explorer',
          query: {
            nontargeted: String(newVal),
            page: String(this.pageNumber),
            sort: this.sortOrder
          }
        })
      }
    },
    defaultPageLink() {
      return {
        name: 'Explorer',
        query: { nontargeted: this.targetedOption, sort: this.sortOrder }
      }
    }
  },
  watch: {
    $route: function() {
      this.getQuestionnaires()
    }
  },
  async created() {
    this.getQuestionnaires()
  },
  methods: {
    getDateStr(str) {
      return common.getDateStr(str)
    },
    getRelativeDateStr(str) {
      return common.relativeDateStr(str)
    },
    getQuestionnaires() {
      this.questionnaires = []
      axios
        .get('/api/admin/questionnaires')
        .then(response => {
          this.questionnaires = response.data.questionnaires
          this.$set(this.range, 'last', response.data.page_max)
        })
        .catch(error => console.log(error))
    },
    changeSortOrder(sortOrder) {
      this.sortOrder = sortOrder
      this.getQuestionnaires()
    },
    changetargetedOption(targetedOption) {
      this.targetedOption = targetedOption
      this.getQuestionnaires()
    }
  }
}
</script>

<style lang="scss" scoped>
td {
  vertical-align: middle;
  font-size: 0.9em;
}
.table-item-title {
  min-width: 10em;
  font-size: 1em;
}
.table-item-date {
  min-width: 8em;
  text-align: center;
}
.dropdowns {
  padding-left: 1.5rem;
  padding-right: 1.5rem;
  padding-top: 1.5rem;
}
.dropdown-content:hover {
  background-color: $base-pink;
}
.dropdown-content.is-selected {
  background-color: $base-brown;
}
.button p {
  margin-right: 0.5em;
}
.card {
  margin: 0;
}
.card-wrapper {
  padding: 1rem 1.5rem;
}
</style>
