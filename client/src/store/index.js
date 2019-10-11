import Vue from 'vue'
import Vuex from 'vuex'
import axios from '@/bin/axios'
import { setAuthToken } from '../utils/api'

Vue.use(Vuex)

const store = new Vuex.Store({
  namespaced: true,
  state: {
    me: null,
    drawer: null,
    color: 'success',
    sidebarBackgroundColor: 'rgba(27, 27, 27, 0.74)',
    loginDialog: false,
    authToken: null,
    cart: []
  },
  mutations: {
    setMe(state, data) {
      state.me = data
    },
    setDrawer(state, data) {
      state.drawer = data
    },
    setColor(state, data) {
      state.color = data
    },
    setToken(state, data) {
      state.authToken = data
      setAuthToken(data)
    },
    toggleDrawer(state) {
      state.drawer = !state.drawer
    },
    toggleLoginDialog(state) {
      state.loginDialog = !state.loginDialog
    },
    item2Cart(state, data) {
      state.cart.push(data)
    },
    removeItemFromCart(state, i) {
      state.cart.splice(i, 1)
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
