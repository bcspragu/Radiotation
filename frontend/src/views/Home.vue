<template>
  <div>
    <div v-if="!user" class="columns is-centered">
      <div class="column is-2">
        <SignIn class="is-large is-fullwidth" @loggedIn="onUserLoggedIn"/>
      </div>
    </div>
    <div v-else>
      <div class="columns is-centered">
        <div class="column is-6 is-10-mobile is-offset-1-mobile instructions">
          <h1 class="is-size-3 has-text-centered">Instructions</h1>
          <ol class="is-size-4">
            <li>Log in with your Google Account.</li>
            <li>Join an existing room with your friends or create a new one.</li>
            <li>Search for your favorite songs, and add them to your playlist.</li>
            <li>Open up the Radiotation app for Android and start playing it back.</li>
          </ol>
          <p class="is-size-5">
            Radiotation will handle the rest, giving everyone equal playtime in the
            car (as long as everyone has added music!)
          </p>
        </div>
      </div>
      <div class="columns is-centered">
        <div class="column is-4">
          <h1 class="has-text-centered is-size-3">Join Room</h1>
          <b-field grouped>
            <b-field expanded label="Room Code">
              <b-input
                autocomplete="off"
                @keyup.native.enter="joinRoom"
                type="text"
                v-model="roomCode"
                name="room-code"
                placeholder="Room Code"></b-input>
            </b-field>
            <b-field class="align-button" label=".">
              <p class="control">
                <button v-on:click="joinRoom" class="button is-primary">Join</button>
              </p>
            </b-field>
          </b-field>
        </div>
        <div class="column is-4 is-offset-1">
          <h1 class="has-text-centered is-size-3">Search for Room</h1>
          <b-field grouped>
            <b-field expanded label="Search">
              <b-input
                autocomplete="off"
                @keyup.native.enter="searchForRoom"
                type="text"
                v-model="searchTerm"
                name="search-room"
                placeholder="Search"></b-input>
            </b-field>
            <b-field class="align-button" label=".">
              <p class="control">
                <button v-on:click="searchForRoom" class="button is-primary">Search</button>
              </p>
            </b-field>
          </b-field>
        </div>
      </div>
      <div class="columns is-centered">
        <div class="column is-6">
          <h1 class="has-text-centered is-size-3">New Room</h1>
          <RoomForm></RoomForm>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';
import RoomForm from '@/components/RoomForm.vue';
import SignIn from '@/components/SignIn.vue';

@Component({
  components: {
    RoomForm,
    SignIn,
  },
})

export default class Home extends Vue {
  // TODO: Add a real User type instead of '{}'.
  private user: {} | null = null;
  private roomCode = '';
  private searchTerm = '';
  private redirect: string = '';

  private created(): void {
    if (this.$route.query.redirect instanceof Array) {
      this.redirect = this.$route.query.redirect[0];
    } else {
      this.redirect = this.$route.query.redirect;
    }
    this.$emit('updateTitle', 'Radiotation');
    this.fetchUser();
  }

  private fetchUser(): void {
    this.$http.get('user').then((response) => {
      const data = response.data;
      if (!data.Error) {
        this.user = data;
      }
    });
  }

  private onUserLoggedIn(user: any): void {
    if (this.user) {
      if (this.redirect) {
        this.$router.push({path: this.redirect});
      }
      return;
    }

    user.getIdToken().then((token: string) => {
      this.$http.post('verifyToken', {token, name: user.displayName, anonymous: user.isAnonymous}).then(() => {
        this.fetchUser();
      });
    });
  }

  private joinRoom(): void {
    this.$router.push({name: 'Room', params: {id: this.roomCode}});
  }

  private searchForRoom(): void {
    this.$router.push({name: 'RoomSearch', query: {query: this.searchTerm}});
  }

}
</script>

<style scoped>
.instructions {
  margin-top: 1em;
}

#g-signin {
  display: inline-block
}
</style>
