<template>
  <div id="app">
    <nav class="navbar" role="navigation" aria-label="main navigation">
      <div class="navbar-brand">
          <router-link :to="tt.to" v-for="(tt, index) in title" class="is-size-3 logo-title navbar-item">
            <img class="logo" src="./assets/radiotation_logo.png">
            {{ tt.text }}
            <span v-if="index < title.length-1"> &gt; </span>
          </router-link>
        </span>
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
      title: [{text: 'Radiotation', to: { name: 'Home' }}],
    }
  },
  methods: {
    updateTitle (title) {
      // this.title = title;
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
