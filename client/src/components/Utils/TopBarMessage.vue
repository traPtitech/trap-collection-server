<template>
  <transition name="fade">
    <div
      v-if="message.showMessage"
      class="message"
      :class="{
        'is-danger': message.color === 'red',
        'is-success': message.color === 'green'
      }"
    >
      <p class="message-body">{{ message.body }}</p>
    </div>
  </transition>
</template>

<script>
export default {
  name: 'TopBarMessage',
  components: {},
  props: {
    message: {
      type: Object,
      required: true
    }
  },
  data() {
    return {}
  },
  computed: {},
  mounted() {},
  methods: {}
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.message {
  margin: 0;
}
.message-body {
  font-size: 1rem;
  // margin: 1rem 0 1rem auto;
  margin: 1rem 0.75rem;
  padding: 0.5rem 0.75rem;
  width: fit-content;
  // border-radius: 4px;
  border-style: solid;
  border-width: 0 0 0 4px;
}
</style>
