<template>
  <div row justify-center>
    <div v-if="!loginDialog" persistent max-width="290">
      <div>
        <div class="headline">
          ログインしてください
        </div>
        <div>
          OKを押すとtraQに飛びます。<br />
          「承認」を押すとログインできます。
        </div>
        <div>
          <div></div>
          <button color="green darken-1" @click="login">OK</button>
        </div>
      </div>
    </div>
  </div>
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
