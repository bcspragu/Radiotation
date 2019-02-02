<template>
  <div>
    <div id="firebaseui-auth-container"></div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';

import 'firebase/auth';
import * as firebase from 'firebase/app';
import * as firebaseui from 'firebaseui';

@Component
export default class SignIn extends Vue {
  private mounted(): void {
    const ui = new firebaseui.auth.AuthUI(firebase.auth());
    ui.start('#firebaseui-auth-container', {
      signInOptions: [
        firebase.auth.EmailAuthProvider.PROVIDER_ID,
        firebase.auth.GoogleAuthProvider.PROVIDER_ID,
        firebaseui.auth.AnonymousAuthProvider.PROVIDER_ID,
      ],
      // Other config options...
    });
  }
}
</script>

