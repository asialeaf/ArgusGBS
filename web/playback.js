import Vue from 'vue'

import Playback from 'components/Playback.vue'
import store from "./store"

new Vue({
  el: '#app',
  store,
  template: `<Playback></Playback>`,
  components: {
    Playback
  }
})