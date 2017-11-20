<template>
  <div id="app">
    <header class="navbar">
      <section class="navbar-section">
        <img class="logo" src="assets/radiotation_logo.png">
        <div class="btn-home btn btn-link">{{title}}</div>
      </section>
      <section class="navbar-section"></section>
    </header>
    <router-view v-on:updateTitle="updateTitle"/>
  </div>
</template>

<script>
export default {
  name: 'app',
  data () {
    return {
      title: 'Radiotation'
    }
  },
  methods: {
    updateTitle (title) {
      this.title = title
    }
  }
}
</script>

<style scoped>
.navbar {
  background: #F8F9FA;
}

.logo {
  width:auto;
  height:auto;
  max-height:60px;
}

.btn-home {
  font-size: 30px;
}
</style>
