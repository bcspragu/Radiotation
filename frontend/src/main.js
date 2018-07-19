import Vue from 'vue'
import App from './App.vue'
import router from './router'

import VueResource from 'vue-resource'

Vue.use(VueResource)

Vue.config.productionTip = false
Vue.http.options.root = '/api';

import Buefy from 'buefy'

Vue.component(Buefy.Field.name, Buefy.Field);
Vue.component(Buefy.Input.name, Buefy.Input);
Buefy.install(Vue);

new Vue({
  router: router,
  render: h => h(App),
  http: {
    root: '/api',
  }
}).$mount('#app')
