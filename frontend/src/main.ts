import Vue from 'vue';
import Buefy from 'buefy';
import 'buefy/dist/buefy.css';
import 'firebaseui/dist/firebaseui.css';
import axios from 'axios';

import App from './App.vue';
import router from './router';

Vue.config.productionTip = false;

Vue.use(Buefy);

Vue.prototype.$http = axios.create({
  baseURL: '/api',
});

new Vue({
  router,
  render: (h) => h(App),
}).$mount('#app');
