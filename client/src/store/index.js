import Vue from 'vue'
import Vuex from 'vuex'
import axios from '@/bin/axios'

Vue.use(Vuex)

const store = new Vuex.Store({
  namespaced: true,
  state: {
    me: null
  },
  getters: {
    getMe(state) {
      return state.me
    },
    getMyTraqId(state) {
      return state.me !== null ? state.me.traqId : ''
    }
  },
  mutations: {
    setMe(state, me) {
      state.me = me
    },
    setMyTraqId(state, traqId) {
      if (!state.me) state.me = {}
      state.me.traqId = traqId
    }
  },
  actions: {
    whoAmI({ commit }) {
      return axios
        .get('/trap/users/me')
        .then(res => {
          if (res.data.traqID !== '-') {
            // traQにログイン済みの場合
            commit('setMe', { traqId: res.data.traqID })
          }
        })
        .catch(err => {
          console.log(err)
        })
    }
  }
})

export default store
