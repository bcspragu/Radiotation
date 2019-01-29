<template>
  <h1>Welcome to My Awesome App</h1>
  <div id="firebaseui-auth-container"></div>
  <div id="loader">Loading...</div>
</template>

<script>
var firebase = require('firebase');
var firebaseui = require('firebaseui');

export default {
  name: 'SignIn',
  data () {
    return {}
  },
  mounted () {
    // Initialize Firebase
    var config = {
      apiKey: "AIzaSyDuNshEtUFNM7bZAl62aQWyp_EBgUZGvaQ",
      authDomain: "radiotation-169318.firebaseapp.com",
      databaseURL: "https://radiotation-169318.firebaseio.com",
      projectId: "radiotation-169318",
      storageBucket: "radiotation-169318.appspot.com",
      messagingSenderId: "990909845123"
    };
    firebase.initializeApp(config);

    var ui = new firebaseui.auth.AuthUI(firebase.auth());
    ui.start('#firebaseui-auth-container', {
      signInOptions: [
        firebase.auth.EmailAuthProvider.PROVIDER_ID,
        firebase.auth.GoogleAuthProvider.PROVIDER_ID,
        firebase.auth.FacebookAuthProvider.PROVIDER_ID,
      ],
      // Other config options...
    });
  }
}
</script>
