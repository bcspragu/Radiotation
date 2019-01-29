import Vue from 'vue';
import Router from 'vue-router';

import Home from './views/Home.vue';
import RoomSearch from './views/RoomSearch.vue';
import Room from './views/Room.vue';
import SongSearch from './views/SongSearch.vue';
import CreateRoom from './views/CreateRoom.vue';

Vue.use(Router);

export default new Router({
  mode: 'history',
  base: process.env.BASE_URL,
  routes: [
    {
      path: '/',
      name: 'Home',
      component: Home,
    },
    {
      path: '/search',
      name: 'RoomSearch',
      component: RoomSearch,
    },
    {
      path: '/room/:id',
      name: 'Room',
      component: Room,
    },
    {
      path: '/room/:id/search',
      name: 'SongSearch',
      component: SongSearch,
    },
    {
      path: '/room/:id/create',
      name: 'CreateRoom',
      component: CreateRoom,
    },
  ],
});
