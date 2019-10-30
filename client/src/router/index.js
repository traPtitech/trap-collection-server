import Vue from 'vue'
import store from '@/store'
import Router from 'vue-router'
import Explorer from '@/pages/Explorer'
import QuestionnaireDetails from '@/pages/QuestionnaireDetails'
import Results from '@/pages/Results'
import Seat from '@/pages/Seat'
import NotFound from '@/pages/NotFound'
import { fetchAuthToken, setAuthToken, getMe } from '../utils/api'

setAuthToken(store.state.authToken)

Vue.use(Router)

const router = new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      redirect: '/explorer'
    },
    {
      path: '/explorer',
      name: 'Explorer',
      component: Explorer
    },
    {
      path: '/questionnaires/:id',
      name: 'QuestionnaireDetails',
      component: QuestionnaireDetails
    },
    {
      path: '/results/:id',
      name: 'Results',
      component: Results
    },
    {
      path: '/seat',
      name: 'Seat',
      component: Seat
    },
    {
      path: '*',
      name: 'NotFound',
      component: NotFound
    },
    {
      path: '/callback',
      name: 'callback',
      component: () => import('../components/Home.vue'),
      beforeEnter: async (to, from, next) => {
        const code = to.query.code
        const state = to.query.state
        const codeVerifier = sessionStorage.getItem(
          `login-code-verifier-${state}`
        )
        if (!code || !codeVerifier) {
          next('/')
        }
        try {
          const res = await fetchAuthToken(code, codeVerifier)
          await store.commit('setToken', res.data.access_token)
          await setAuthToken(res.data.access_token)
          store.commit('toggleLoginDialog')
          const resp = await getMe()
          await store.commit('setMe', resp.data)
          next('/')
        } catch (e) {
          // eslint-disable-next-line no-console
          console.error(e)
        }
      }
    }
  ],
  scrollBehavior(savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      // ページ遷移の時ページスクロールをトップに
      return { x: 0, y: 0 }
    }
  }
})

export default router
