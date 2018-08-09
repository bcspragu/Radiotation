<template>
  <div id="app">
    <nav class="navbar" role="navigation" aria-label="main navigation">
      <div class="navbar-brand">
        <div v-for="(tt, index) in title">
          <router-link :to="tt.to" class="is-size-5 navbar-item">
            <img v-if="index == 0" class="logo" src="./assets/radiotation_logo.png">
            {{ tt.text }}
          </router-link>
        </div>
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
    updateTitle (mod) {
      switch (mod.mod) {
        case 'add':
          this.title.push(mod.item)
          break
        case 'set':
          this.title = mod.item
        case 'pop':
          this.title.pop()
      }
      console.log(this.title)
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
