<template>
  <div class="container">
    <div class="col-mx-auto col-6 column instructions">
      <h3>Instructions</h3>
      <ol>
        <li>Log in with you Google Account.</li>
        <li>Join an existing room with your friends or create a new one.</li>
        <li>Search for your favorite songs, and add them to your playlist.</li>
        <li>Open up the Radiotation app for Android and start playing it back.</li>
      </ol>
      <p>
        Radiotation will handle the rest, giving everyone equal playtime in the
        car (as long as everyone has added music!)
      </p>
    </div>
    <div v-if="user" class="columns">
      <div class="column col-6 col-sm-12">
        <h2 class="text-center">New Room</h2>
        <room-form></room-form>
      </div>
      <div class="column col-6 col-sm-12">
        <h2 class="text-center">Available Rooms</h2>
        <ul>
          <li v-for="room in rooms">{{ room }}</li>
        </ul>
      </div>
    </div>
    <div v-else class="columns signin-holder">
      <div>
        <div id="g-signin"></div>
      </div>
    </div>
  </div>
</template>

<script>
import RoomForm from './RoomForm.vue'

export default {
  name: 'Home',
  data () {
    return {
      user: null,
      rooms: []
    }
  },
  components: {
    'room-form': RoomForm
  },
  created () {
    this.fetchUser()
  },
  methods: {
    fetchUser () {
      var vue = this
      vue.$http.get('/user').then(response => {
        var data = JSON.parse(response.body)
        if (data.Error) {
          // eslint-disable-next-line
          gapi.signin2.render('g-signin', {
            'scope': 'profile email',
            'width': 240,
            'height': 50,
            'theme': 'dark',
            'onsuccess': vue.onSignIn,
            'onfailure': vue.onFailure
          })
        } else {
          vue.user = data
        }
      })
    },
    onSignIn (googleUser) {
      if (this.user) {
        return
      }
      var data = {token: googleUser.getAuthResponse().id_token}
      this.$http.post('/verifyToken', data, {emulateJSON: true}).then(this.fetchUser)
    }
  }
}
</script>

<style scoped>
.instructions {
  margin-top: 2em;
}
</style>
