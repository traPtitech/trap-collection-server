<template>
  <nav class="navbar is-fixed-bottom is-flex">
    <div class="editor-buttons">
      <slot></slot>
    </div>
  </nav>
</template>

<script>
export default {
  name: 'EditNavBar',
  components: {},
  props: {},
  data() {
    return {}
  },
  computed: {},
  mounted() {},
  methods: {}
}
</script>

<style lang="scss" scoped>
.editor-buttons {
  display: inline-flex;
  margin: 0 0 0 auto;
  .button {
    margin: 0;
    border-radius: 0;
    height: auto;
    border-width: 3px 0;
    span,
    span:before {
      line-height: normal;
      font-weight: bold;
      vertical-align: sub;
    }
    &.save-button {
      width: 4rem;
      background-color: $button-background-color-gray;
      border-color: $button-border-color-gray;
    }
    &.cancel-button {
      width: 4rem;
      background-color: $button-background-color-red;
      border-color: $button-border-color-red;
    }
    &.send-button {
      background-color: $button-background-color-green;
      border-color: $button-border-color-green;
      min-width: 10rem;
      [class^='ti-'] {
        margin: auto 0.3em;
      }
    }
    &[disabled] {
      background-color: #c5c5c5;
      border-color: #999999;
      color: #999999;
    }
    &:hover {
      border-color: $base-pink;
    }
  }
  @media screen and (max-width: 768px) {
    width: -webkit-fill-available;
    .button {
      width: -webkit-fill-available;
    }
  }
}
.navbar.is-fixed-bottom {
  background-color: $var-indigo;
}
</style>
