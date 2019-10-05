<template>
  <div v-show="inputError.isError" class="message is-warning">
    <p class="message-header error-message">{{ inputError.message }}</p>
  </div>
</template>

<script>
// import <componentname> from '<path to component file>'

export default {
  name: 'InputErrorMessage',
  components: {},
  props: {
    inputError: {
      type: Object,
      required: true
    }
  },
  data() {
    return {}
  }
}
</script>

<!-- Add 'scoped' attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.message {
  margin-bottom: 0.5rem;
  .error-message {
    font-size: 1rem;
    margin: 0.5rem;
  }
}
</style>
