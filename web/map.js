import Vue from 'vue'

import MapPage from 'components/MapPage.vue'
import store from "./store"

new Vue({
  el: '#app',
  store,
  template: `<MapPage></MapPage>`,
  components: {
    MapPage
  }
})