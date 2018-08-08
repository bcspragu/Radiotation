import Vue from 'vue'
import Router from 'vue-router'

import CreateRoom from '@/components/CreateRoom'
import Home from '@/components/Home'
import Room from '@/components/Room'
import RoomList from '@/components/RoomList'
import Search from '@/components/Search'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'Home',
      component: Home
    },
    {
      path: '/search',
      name: 'RoomList',
      component: RoomList
    },
    {
      path: '/room/:id',
      name: 'Room',
      component: Room
    },
    {
      path: '/room/:id/search',
      name: 'Search',
      component: Search
    },
    {
      path: '/room/:id/create',
      name: 'CreateRoom',
      component: CreateRoom
    }
  ]
})
