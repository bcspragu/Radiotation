import Vue from 'vue';
import Buefy from 'buefy';
import 'buefy/dist/buefy.css';
import 'firebaseui/dist/firebaseui.css';
import { User } from 'firebase';
import axios from 'axios';

import App from './App.vue';
import router from './router';
import { RoomInfo } from '@/data';

Vue.config.productionTip = false;

Vue.use(Buefy);

Vue.prototype.$http = axios.create({
  baseURL: '/api',
});

interface RoomResults {
  displayName: string;
  roomCode: string;
  numberUsers: number;
}

class Store {
  public firebaseUser: User | null = null;
  public loggedIn: boolean = false;
  public searchResults: RoomResults[] = [];
  public roomInfo: RoomInfo | null = null;

}
const store = new Store();

new Vue({
  router,
  render: (h) => h(App),
  data: store,
}).$mount('#app');
