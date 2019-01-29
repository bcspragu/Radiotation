<template>
  <div id="app">
    <nav class="navbar" role="navigation" aria-label="main navigation">
      <div class="navbar-brand">
        <router-link class="navbar-item" :to="{ name: 'Home' }">
          <img src="./assets/radiotation_logo.png" alt="Radiotation Logo" width="28" height="28">
        </router-link>
      </div>
      <div class="is-size-5 navbar-item">
        {{ title }}
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
      title: '',
    }
  },
  methods: {
    updateTitle (title) {
      this.title = title
    },
    handleError (data) {
      // eslint-disable-next-line
      console.log(data);
    }
  }
}
</script>

<style scoped>
</style>

<style lang="scss">
html {
  overflow-y: auto !important;
}

html, body, #app {
  height: 100%;
  margin: 0;
}

#app {
  display: flex;
  flex-flow: column;
  height: 100%;
}

nav {
  box-shadow: 0 4px 4px -4px #000000;
  margin-bottom: 1rem;
}

// Import Bulma's core
@import "~bulma/sass/utilities/_all";

// Import Bulma and Buefy styles
@import "~bulma";
@import "~buefy/src/scss/buefy";
</style>
