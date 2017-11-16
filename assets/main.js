import Vue from 'vue';
import VueRouter from 'vue-router';

import App from './App.vue';
import Home from './Home.vue';
import Room from './Room.vue';
import '../node_modules/spectre.css/dist/spectre.min.css'

//Vue.use(VueRouter);

var routes = [
  { path: '/', component: Home },
  { path: '/room/:id', component: Room }
]

const router = new VueRouter({ routes: routes })

new Vue({
  el: '#app',
  //router: router,
  template: '<Home/>',
  components: { Home: Home }
});

var conn;

$(function() {
  if (window["WebSocket"]) {
    conn = new WebSocket("wss://localhost:8000/ws");
    conn.onclose = function(evt) {
      // Something
    }
    conn.onmessage = function(evt) {
    }
  } else {
    // You ain't got WebSockets, brah
  }
});
