<template>
  <div id="app">
    <header class="navbar">
      <section class="navbar-section navbar-main">
        <img class="logo" src="static/radiotation_logo.png">
        <div class="btn-header btn btn-link">{{title}}</div>
      </section>
      <section class="navbar-section"></section>
    </header>
    <router-view v-on:updateTitle="updateTitle" v-on:ajaxErr="handleError"/>
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
    },
    handleError (data) {
      if (data && data.NotLoggedIn) {
        this.$router.push({name: 'Home'})
        return
      }
      console.log(data)
    }
  }
}
</script>

<style>
html, body, #app {
  height: 100%;
}
</style>

<style scoped>
#app {
  display: flex;
  flex-direction: column;
}

.navbar {
  background: #F8F9FA;
  display: block;
}

.navbar-main {
  max-width: 100%;
}

.logo {
  width:auto;
  height:auto;
  max-height:40px;
}

.btn-header {
  font-size: 16px;
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}
</style>
