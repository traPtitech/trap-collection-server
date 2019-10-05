module.exports = {
  root: true,
  parser: 'vue-eslint-parser',
  parserOptions: {
    parser: 'babel-eslint',
    sourceTye: 'module'
  },
  env: {
    browser: true,
    node: true
  },
  extends: [
    'eslint:recommended',
    'plugin:prettier/recommended',
    'plugin:vue/recommended',
    'prettier/vue'
  ],
  rules: {
    'no-plusplus': 'off',
    'no-console': 'off',
    'func-names': 'off',
    'vue/no-template-shadow': 0,
    'vue/component-name-in-template-casing': 1, // <template> にコンポーネントを書く時はkebab-case
    'prettier/prettier': [
      'error',
      {
        singleQuote: true,
        semi: false,
        bracketSpacing: true,
        tabWidth: 2
      }
    ]
  }
}
