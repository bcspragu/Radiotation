<template>
  <div>
    <div class="columns is-centered is-mobile">
      <div class="column is-6 instructions">
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
    <div class="columns">
      <div class="column is-12 is-6-mobile">
        <h1 class="has-text-centered is-size-3">Join Room</h1>
        <div class="form-horizontal">
          <div class="form-group">
            <div class="col-3">
              <label for="room" class="form-label">Room Code</label>
            </div>
            <div class="col-9">
              <input 
                autocomplete="off"
                v-on:keyup.enter="joinRoom"
                type="text"
                v-model="roomCode"
                name="room-code"
                class="form-input"
                placeholder="Room Code">
            </div>
          </div>
        </div>
      </div>
      <div class="column col-6 col-sm-12">
        <h2 class="text-center">New Room</h2>
        <room-form></room-form>
      </div>
      <sign-in-button @done="onUserLoggedIn"/>
    </div>
  </div>
</template>

<script>
import RoomForm from '@/components/RoomForm.vue'
import SignIn from '@/components/SignIn.vue'

export default {
  name: 'Home',
  data () {
    return {
      user: null,
      roomCode: '',
      redirect: this.$route.query.redirect
    }
  },
  components: {
    'room-form': RoomForm,
    'sign-in-button': SignIn
  },
  created () {
    this.$emit('updateTitle', 'Radiotation')
    this.fetchUser()
  },
  methods: {
    fetchUser () {
      var vue = this
      vue.$http.get('user').then(response => {
        var data = response.body;
        if (!data.Error) {
          vue.user = data
        }
      })
    },
    onUserLoggedIn (googleUser) {
      if (this.user) {
        if (this.redirect) {
          this.$router.push({path: this.redirect})
        }
        return
      }
      var data = {token: googleUser.getAuthResponse().id_token}
      this.$http.post('/verifyToken', data, {emulateJSON: true}).then(() => {
        if (this.redirect) {
          this.$router.push({path: this.redirect})
          return
        }
        this.fetchUser()
      })
    },
    joinRoom () {
      this.$router.push({name: 'Room', params: {id: this.roomCode}})
    }
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
