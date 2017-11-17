<template>
  <div class="form-horizontal">
    <div class="form-group">
      <div class="col-3">
        <label for="room" class="form-label">Room Name</label>
      </div>
      <div class="col-9">
        <input type="text" v-model="roomName" name="room" class="form-input" placeholder="Room Name" value="">
      </div>
    </div>
    <div class="form-group">
      <div class="col-3">
        <label for="musicSource" class="form-label">Music Source</label>
      </div>
      <div class="col-9">
        <select v-model="musicSource" name="musicSource" class="form-select">
          <!--<option value="playmusic">Google Play Music</option>-->
          <option value="spotify">Spotify</option>
        </select>
      </div>
    </div>
    <div class="form-group">
      <div class="col-3">
        <label for="shuffleOrder" class="form-label">Shuffle Order</label>
      </div>
      <div class="col-9">
        <select v-model="shuffleOrder" name="shuffleOrder" class="form-select">
          <option value="robin">Round Robin</option>
          <option value="shuffle">Fair Random</option>
          <option value="random">True Random</option>
        </select>
      </div>
    </div>
    <div class="form-group">
      <button v-on:click="createRoom" class="btn btn-lg centered">Create Room</button>
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
