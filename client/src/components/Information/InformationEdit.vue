<template>
  <div>
    <form @submit.prevent>
      <div class="columns is-flex">
        <article class="column is-11">
          <div class="card">
            <!-- タイトル、説明、回答期限 -->
            <div>
              <header class="card-header">
                <div id="title" class="card-header-title title is-editing">
                  <div class="wrapper">
                    <input
                      :value="information.title"
                      class="input"
                      placeholder="タイトル"
                      @input="$set(information, 'title', $event.target.value)"
                    />
                  </div>
                </div>
              </header>
              <input-error-message
                :input-error="inputErrors.noTitle"
              ></input-error-message>
              <div class="card-content">
                <textarea
                  id="description"
                  v-model="information.description"
                  class="textarea"
                  rows="5"
                  placeholder="説明"
                ></textarea>
              </div>
              <div class="is-pulled-right is-inline-block wrapper">
                <div class="wrapper editable">
                  <span class="label">回答期限 :</span>
                  <input
                    v-model="resTimeLimitEditStr"
                    class="input"
                    type="datetime-local"
                    :disabled="noTimeLimit"
                  />
                </div>
                <label class="checkbox is-pulled-right">
                  <input v-model="noTimeLimit" type="checkbox" />
                  期限なし
                </label>
              </div>
            </div>
          </div>
        </article>
      </div>

      <article class="column is-5">
        <div class="card">
          <!-- 操作 -->
          <div>
            <header class="card-header">
              <div class="card-header-title subtitle">操作</div>
            </header>
            <div class="card-content management-buttons">
              <management-button
                :questionnaire-id="questionnaireId"
                type="deleteQuestionnaire"
              ></management-button>
            </div>
          </div>
        </div>
      </article>
    </form>
  </div>
</template>

<script>
import common from '@/bin/common'
import axios from '@/bin/axios'
import InputErrorMessage from '@/components/Utils/InputErrorMessage'
import ManagementButton from '@/components/Information/ManagementButton'

export default {
  name: 'InformationEdit',
  components: {
    'input-error-message': InputErrorMessage,
    'management-button': ManagementButton
  },
  props: {
    informationProps: {
      type: Object,
      required: true
    },
    inputErrors: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      responses: [],
      activeModal: {},
      isModalActive: false,
      users: {},
      groupTypes: {}
      // usersIsSelected: {}
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
    noTimeLimit: {
      get() {
        return this.informationProps.noTimeLimit
      },
      set(newBool) {
        this.$emit('set-data', 'noTimeLimit', newBool)
      }
    },
    isNewQuestionnaire() {
      return this.$route.params.id === 'new'
    },
    userLists() {
      return common.getUserLists(
        this.information.targets,
        this.information.respondents,
        this.information.administrators
      )
    },
    resSharedToStr: {
      get: function() {
        switch (this.information.res_shared_to) {
          case 'public':
            return '全体'
          case 'administrators':
            return '管理者のみ'
          case 'respondents':
            return '回答済みの人'
          default:
            console.error('unexpected res_shared_to')
            return null
        }
      },
      set: function(str) {
        switch (str) {
          case '全体': {
            this.information.res_shared_to = 'public'
            break
          }
          case '管理者のみ': {
            this.information.res_shared_to = 'administrators'
            break
          }
          case '回答済みの人': {
            this.information.res_shared_to = 'respondents'
            break
          }
        }
      }
    },
    resTimeLimitEditStr: {
      get: function() {
        if (
          !this.information.res_time_limit ||
          this.information.res_time_limit === 'NULL'
        )
          return ''
        return this.information.res_time_limit.slice(0, 16)
      },
      set: function(str) {
        if (str === '') {
          this.$emit('set-data', 'noTimeLimit', true)
        } else {
          this.information.res_time_limit = str
        }
      }
    },
    targetedList: {
      get() {
        return this.information.targets
      },
      set(newVal) {
        this.setUserList('targets', newVal)
      }
    }
  },
  watch: {},
  created() {
    // this.getUsers()
    // .then(this.getGroupTypes)
  },
  mounted() {},
  methods: {
    getDateStr(str) {
      return common.getDateStr(str)
    },
    setInformation(newInformation) {
      this.$emit('set-data', 'information', newInformation)
    },
    changeActiveModal(obj) {
      this.activeModal = obj
      this.isModalActive = true
    },
    disableModal() {
      this.isModalActive = false
    },
    setUserList(listName, newList) {
      let newInformation = this.information
      newInformation[listName] = newList
      this.setInformation(newInformation)
    },
    getUsers() {
      return axios
        .get('https://q.trap.jp/api/1.0/users')
        .then(res => {
          res.data.forEach(user => {
            if (user.accountStatus === 1) {
              this.users[user.userId] = user
            }
          })
        })
        .catch(err => {
          console.log(err)
        })
    },
    getGroupTypes() {
      return axios.get('https://q.trap.jp/api/1.0/groups').then(res => {
        let tmp = {}
        res.data.forEach(group => {
          if (typeof tmp[group.type] === 'undefined') {
            tmp[group.type] = []
          }
          // 除名されていないメンバーをtraQID順にソートしたtraQIDのリストactiveMembersを作成
          group.activeMembers = group.members
            .filter(
              userId =>
                typeof this.users[userId] !== 'undefined' &&
                this.users[userId].accountStatus === 1 &&
                this.users[userId].name !== 'traP'
            )
            .map(userId => this.users[userId].name)
            .sort((a, b) => {
              return a.toLowerCase().localeCompare(b.toLowerCase())
            })
          tmp[group.type].push(group)
        })

        // typeごとに、group名をソートしたものをgroupTypesに入れる
        Object.keys(tmp).forEach(type => {
          this.$set(this.groupTypes, type, {})
          tmp[type].sort((a, b) => {
            return a.name.toLowerCase().localeCompare(b.name.toLowerCase())
          })
        })
        this.groupTypes = tmp
      })
    }
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.card-header-title.is-editing {
  padding: 0;
}
.editable {
  span {
    width: fit-content;
    height: fit-content;
    top: 0;
    bottom: 0;
    margin: auto 0.2rem auto 0;
    white-space: nowrap;
  }
}
.editable.wrapper {
  display: flex;
}
.details {
  .checkbox {
    width: 6rem;
    margin: 0.5rem;
  }
}
.management-buttons {
  .button:not(:last-child) {
    margin-bottom: 0.7rem;
  }
  .button {
    max-width: fit-content;
    display: block;
  }
}
#title {
  .input {
    font-size: 2rem;
  }
  .wrapper {
    width: 100%;
  }
}
.editor-buttons {
  margin: auto;
  .button {
    margin: 1rem;
    // margin-bottom: 1rem;
    width: 8rem;
    max-width: 100%;
  }
}
.user-list-wrapper {
  margin-top: 0.5rem;
}
.user-list {
  margin: 0.2rem 0.5rem;
  label {
    margin: auto 0.2rem;
  }
}
.targets-description {
  color: gray;
  margin: auto 0.5rem;
}
</style>
