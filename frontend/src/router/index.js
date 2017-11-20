import Vue from 'vue'
import Router from 'vue-router'

import Home from '@/components/Home'
import Room from '@/components/Room'
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
      path: '/room/:id',
      name: 'Room',
      component: Room
    },
    {
      path: '/room/:id/search',
      name: 'Search',
      component: Search
    }
  ]
})
