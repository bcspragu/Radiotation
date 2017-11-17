<template>
  <div class="container">
    <div class="row">
      <div class="eight columns offset-by-two">
        <label for="room">Room Name</label>
        <input type="text" v-model="roomName" name="room" class="room-name u-full-width" placeholder="Room Name" value="">
      </div>
    </div>
    <div class="row">
      <div class="four columns offset-by-two">
        <label for="musicSource">Music Source</label>
        <select v-model="musicSource" name="musicSource" class="u-full-width">
          <!--<option value="playmusic">Google Play Music</option>-->
          <option value="spotify">Spotify</option>
        </select>
      </div>
      <div class="four columns">
        <label for="shuffleOrder">Shuffle Order</label>
        <select v-model="shuffleOrder" name="shuffleOrder" class="u-full-width">
          <option value="robin">Round Robin</option>
          <option value="shuffle">Fair Random</option>
          <option value="random">True Random</option>
        </select>
      </div>
    </div>

    <div class="row">
      <button v-on:click="createRoom" class="button eight columns offset-by-two">Create Room</button>
    </div>
  </div>
</template>

<script>
export default {
  data () {
    return {
      roomName: '',
      musicSource: 'spotify',
      shuffleOrder: 'robin'
    }
  },
  methods: {
    createRoom () {
      var data = {
        roomName: this.roomName,
        musicSource: this.musicSource,
        shuffleOrder: this.shuffleOrder
      }
      var vue = this
      vue.$http.post('/room', data, {emulateJSON: true}).then(response => {
        var data = JSON.parse(response.body)
        if (data.Error) {
          // TODO: Handle error
          console.log(data.Error)
          return
        }
        vue.$router.push({name: 'Room', params: {id: data.ID}})
      })
    }
  }
}
</script>
