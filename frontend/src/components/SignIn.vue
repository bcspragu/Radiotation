<template>
  <div>
    <div id="firebaseui-auth-container"></div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';
import * as firebase from 'firebase/app';
import 'firebase/auth';
import * as firebaseui from 'firebaseui';

@Component
export default class SignIn extends Vue {
  private mounted(): void {
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

    const ui = new firebaseui.auth.AuthUI(firebase.auth());
    ui.start('#firebaseui-auth-container', {
      signInOptions: [
        firebase.auth.EmailAuthProvider.PROVIDER_ID,
        firebase.auth.GoogleAuthProvider.PROVIDER_ID,
        firebaseui.auth.AnonymousAuthProvider.PROVIDER_ID,
      ],
      // Other config options...
    });

    firebase.auth().onAuthStateChanged((user) => {
      if (!user) {
        this.$router.push({name: 'Home'});
        return;
      }
      this.$emit('loggedIn', user);
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
}
</script>

