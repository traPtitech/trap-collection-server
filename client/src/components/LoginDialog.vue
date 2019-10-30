<template>
  <el-dialog
    :visible.sync="loginDialog"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="false"
    title="ログインしてください"
  >
    <div>
      OKを押すとtraQに飛びます。<br />
      「承認」を押すとログインできます。
    </div>
    <el-button type="primary" @click="login">OK</el-button>
  </el-dialog>
</template>

<script>
import { mapMutations, mapState } from 'vuex'
import { redirectAuthorizationEndpoint } from '@/utils/api'
export default {
  computed: {
    ...mapState(['loginDialog'])
  },
  methods: {
    ...mapMutations(['toggleLoginDialog']),
    login() {
      this.toggleLoginDialog()
      redirectAuthorizationEndpoint()
    }
  }
}
</script>
