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
        <h2 class="text-center">Available Rooms</h2>
        <div class="text-center available-room" v-for="room in rooms">
          <router-link :to="{ name: 'Room', params: { id: room.ID }}">{{room.DisplayName}}</router-link>
        </div>
      </div>
      <div class="column col-6 col-sm-12">
        <h2 class="text-center">New Room</h2>
        <room-form></room-form>
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
      rooms: [],
      redirect: this.$route.query.redirect
    }
  },
  components: {
    'room-form': RoomForm
  },
  created () {
    this.$emit('updateTitle', 'Radiotation')
    this.fetchUser()
    this.fetchRooms()
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
    fetchRooms () {
      this.$http.get(`/rooms`).then(response => {
        var data = JSON.parse(response.body)
        if (data.Error) {
          this.$emit('ajaxErr', data)
          return
        }
        this.rooms = data
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
    }
  }
}
</script>

<style scoped>
.instructions {
  margin-top: 1em;
}

.available-room {
  font-size: 24px;
}
</style>
