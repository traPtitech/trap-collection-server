import Vue from 'vue'
import store from '@/store'
import Router from 'vue-router'
import Explorer from '@/pages/Explorer'
import QuestionnaireDetails from '@/pages/QuestionnaireDetails'
import Results from '@/pages/Results'
import Seat from '@/pages/Seat'
import NotFound from '@/pages/NotFound'

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

router.beforeEach(async (to, _, next) => {
  // traQにログイン済みかどうか調べる
  if (!store.state.me) {
    await store.dispatch('whoAmI')
  }

  if (!store.state.me) {
    // 未ログインの場合、traQのログインページに飛ばす
    const traQLoginURL = 'https://q.trap.jp/login?redirect=' + location.href
    location.href = traQLoginURL
  }

  next()
})

export default router
