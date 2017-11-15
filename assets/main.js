import Vue from 'vue';
import Home from './Home.vue';
import RoomForm from './RoomForm.vue';
import '../node_modules/spectre.css/dist/spectre.min.css'

Vue.component('room-form', RoomForm);

new Vue({
  el: '#app',
  render: h => h(Home)
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
