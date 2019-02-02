<template>
  <div id="app">
  <nav class="navbar" role="navigation" aria-label="main navigation">
    <div class="navbar-brand">
      <router-link :to="{name: 'Home'}" class="navbar-item" href="https://bulma.io">
        <img src="./assets/radiotation_logo.png" width="28" height="28">
      </router-link>

      <a role="button" class="navbar-burger burger" @click="toggleMenu" aria-label="menu" aria-expanded="false">
        <span aria-hidden="true"></span>
        <span aria-hidden="true"></span>
        <span aria-hidden="true"></span>
      </a>
    </div>

    <div class="navbar-menu" :class="{'is-active': showMenu}">
      <div class="navbar-end">
        <div class="navbar-item">
        <div class="buttons">
          <router-link v-if="showSignIn" class="button is-light" :to="{name: 'SignIn'}">
            <strong>Sign in</strong>
          </router-link>
        </div>
      </div>
    </div>
  </div>
</nav>
    <router-view v-on:ajaxErr="handleError"/>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';
import { User } from 'firebase';

import 'firebase/auth';
import * as firebase from 'firebase/app';
import * as firebaseui from 'firebaseui';

@Component
export default class App extends Vue {
  private showMenu = false;

  private created(): void {
    // Initialize Firebase
    const config = {
      apiKey: 'AIzaSyDuNshEtUFNM7bZAl62aQWyp_EBgUZGvaQ',
      authDomain: 'radiotation-169318.firebaseapp.com',
      databaseURL: 'https://radiotation-169318.firebaseio.com',
      projectId: 'radiotation-169318',
      storageBucket: 'radiotation-169318.appspot.com',
      messagingSenderId: '990909845123',
    };
    firebase.initializeApp(config);

    firebase.auth().onAuthStateChanged((user) => {
      if (!user) {
        firebase.auth().signInAnonymously().catch((error) => {
          if (error.code === 'auth/operation-not-allowed') {
            console.log('You must enable Anonymous auth in the Firebase Console.');
          } else {
            console.log(error);
          }
        });
        return;
      }

      this.$root.$data.firebaseUser = user;
      user.getIdToken().then((token: string) => {
        const req = {
          token,
          name: user.displayName,
          anonymous: user.isAnonymous,
        };
        this.$http.post('verifyToken', req).then(() => {
          this.fetchUser();
        });
      });
      /*
      const displayName = user.displayName;
      const email = user.email;
      const emailVerified = user.emailVerified;
      const photoURL = user.photoURL;
      const uid = user.uid;
      const phoneNumber = user.phoneNumber;
      const providerData = user.providerData;
      */
    }, (error) => {
      console.log(error);
    });
  }

  private handleError(data: any): void {
    console.log(data);
  }

  // Show the sign in button if we aren't logged in, or if we are logged in,
  // but anonymously.
  get showSignIn(): boolean {
    const user = this.$root.$data.firebaseUser;
    return !this.$root.$data.loggedIn || (user && user.isAnonymous);
  }

  private fetchUser(): void {
    this.$http.get('user').then((response) => {
      const data = response.data;
      if (!data.Error) {
        this.$root.$data.loggedIn = true;
      }
    });
  }

  private toggleMenu(): void {
    this.showMenu = !this.showMenu;
  }
}
</script>

<style>
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
</style>
