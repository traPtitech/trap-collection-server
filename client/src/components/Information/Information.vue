<template>
  <div>
    <!-- <information-summary :details="summaryProps"></information-summary> -->
    <div class="columns details">
      <article class="column is-6">
        <!-- 情報 -->
        <about
          class="card"
          v-bind="{
            resSharedTo: information.res_shared_to,
            administrators: userLists.administrators,
            respondents: userLists.respondents,
            targets: userLists.targets,
            modifiedAt: information.modified_at,
            createdAt: information.created_at
          }"
        ></about>
      </article>

      <article class="column is-5">
        <div class="card">
          <!-- 操作 -->
          <management
            v-bind="{
              questionnaireInformation: information,
              questionnaireId: questionnaireId,
              canViewResults: canViewResults,
              administrates: administrates
            }"
          ></management>
        </div>
      </article>
    </div>
  </div>
</template>

<script>
// import axios from '@/bin/axios'
import router from '@/router'
import common from '@/bin/common'
import About from '@/components/Information/About'
import Management from '@/components/Information/Management'

export default {
  name: 'Information',
  components: {
    about: About,
    management: Management
  },
  props: {
    informationProps: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      hasResponded: false,
      activeModal: {},
      isModalActive: false,
      newQuestionnaire: false,
      userLists: {}
    }
  },
  computed: {
    information() {
      return this.informationProps.information
    },
    administrates() {
      return this.informationProps.administrates
    },
    questionnaireId() {
      return this.informationProps.questionnaireId
    },
    noTimeLimit() {
      return this.informationProps.noTimeLimit
    },
    canViewResults() {
      // 結果をみる権限があるかどうかを返す
      return common.canViewResults(
        this.information,
        this.administrates,
        this.hasResponded
      )
    }
  },
  watch: {
    information: function() {
      this.userLists = common.getUserLists(
        this.information.targets,
        this.information.respondents,
        this.information.administrators
      )
    }
  },
  created() {
    this.userLists = common.getUserLists(
      this.information.targets,
      this.information.respondents,
      this.information.administrators
    )
  },
  mounted() {},
  methods: {
    createResponse() {
      router.push({
        name: 'NewResponseDetails',
        params: { questionnaireId: this.questionnaireId }
      })
    },
    setHasResponded(bool) {
      this.hasResponded = bool
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.modal-card-head {
  .ti-check {
    background-color: darkgrey;
    color: white;
    font-weight: bolder;
    width: 1.5rem;
    height: 1.5rem;
    padding: 0.25rem;
    border-radius: 1rem;
  }
}
</style>
