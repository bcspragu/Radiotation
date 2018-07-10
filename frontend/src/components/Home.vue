<template>
  <div class="container">
    <div class="columns">
      <div class="col-mx-auto col-6 col-sm-10 column instructions">
        <h3>Instructions</h3>
        <ol>
          <li>Log in with your Google Account.</li>
          <li>Join an existing room with your friends or create a new one.</li>
          <li>Search for your favorite songs, and add them to your playlist.</li>
          <li>Open up the Radiotation app for Android and start playing it back.</li>
        </ol>
        <p>
          Radiotation will handle the rest, giving everyone equal playtime in the
          car (as long as everyone has added music!)
        </p>
      </div>
    </div>
    <div v-if="user" class="columns">
      <div v-if="rooms.length > 0" class="column col-6 col-sm-12">
        <h2 class="text-center">Join Room</h2>
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
    </div>
    <div v-else class="text-center">
      <div id="g-signin"></div>
    </div>
  </div>
</template>

<script>
import RoomForm from '@/components/RoomForm.vue'

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
    'room-form': RoomForm
  },
  created () {
    this.$emit('updateTitle', 'Radiotation')
    this.fetchUser()
  },
  methods: {
    fetchUser () {
      var vue = this
      vue.$http.get('user').then(response => {
        var data = JSON.parse(response.body)
        if (data.Error) {
          // eslint-disable-next-line
          /*
          gapi.signin2.render('g-signin', {
            'scope': 'profile email',
            'width': 240,
            'height': 50,
            'onsuccess': vue.onSignIn,
            'onfailure': vue.onFailure
          })
          */
        } else {
          vue.user = data
        }
      })
    },
    onSignIn (googleUser) {
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
