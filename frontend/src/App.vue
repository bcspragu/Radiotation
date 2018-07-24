<template>
  <div id="app">
    <nav class="navbar" role="navigation" aria-label="main navigation">
      <div class="navbar-brand">
        <router-link :to="{ name: 'Home' }" class="navbar-item">
          <img class="logo" src="./assets/radiotation_logo.png">
          <span class="is-size-3 logo-title">Radiotation</span>
        </router-link>
      </div>
    </nav>
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
      this.title = title;
    },
    handleError (data) {
      // eslint-disable-next-line
      console.log(data);
    }
  }
}
</script>

<style scoped>
.logo-title {
  margin-left: 1rem;
}
</style>

<style lang="scss">
html, body, #app {
  height: 100%;
  margin: 0;
}

#app {
  display: flex;
  flex-flow: column;
  height: 100%;
}

// Import Bulma's core
@import "~bulma/sass/utilities/_all";

// Import Bulma and Buefy styles
@import "~bulma";
@import "~buefy/src/scss/buefy";
</style>
