<template>
  <div id="container">
    更新までに5秒のインターバルがあります
    <div id="main">
      <div id="forword"><span>前</span></div>
      <div
        v-for="i in [0, 1, 2, 3, 4, 5]"
        :key="i"
        :class="{ upper: i % 2 === 0, downer: i % 2 === 1 }"
        class="wrapper"
      >
        <div
          v-for="j in [1, 2, 3, 4, 5, 6]"
          :key="j"
          :class="{ selected: isSelected(i * 6 + j) }"
          class="seat"
          @click="onClick(i * 6 + j)"
        />
        <br />
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
export default {
  name: 'Seat',
  prop: {},
  data() {
    return {
      seat: [],
      reload: null
    }
  },
  mounted() {
    this.getSeat()
    this.reload = setInterval(this.getSeat, 5000)
  },
  destroyed() {
    clearInterval(this.reload)
  },
  methods: {
    isSelected(x) {
      let a = this.seat.find(v => {
        return v === x
      })
      return a !== undefined
    },
    getSeat() {
      axios.get('/api/seat').then(res => {
        this.seat = []
        if (res.data !== null) {
          this.seat = res.data.map(v => Number(v))
        }
      })
    },
    onClick(x) {
      let id = String(x)
      let status
      if (this.isSelected(x)) {
        status = 'out'
      } else {
        status = 'in'
      }
      axios.post('/api/seat', {
        id: id,
        status: status
      })
    }
  }
}
</script>

<style>
.seat {
  width: 15%;
  height: 100pt;
  float: left;
  border: medium solid black;
}
.selected {
  background-color: red;
}
.wrapper {
  margin: auto;
}
.upper {
  margin-top: 10pt;
}
.bottom {
  margin-bottom: 10pt;
}
#forword {
  float: left;
  margin-left: 10pt;
  height: 640pt;
  border: medium solid black;
}
#container {
  margin: 10pt;
}
#main {
  contain: content;
}
</style>
