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
          <!--<option value="gmusic">Google Play Music</option>-->
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
module.exports = {
  data: function() {
    return {
      roomName: "",
      musicSource: "spotify",
      shuffleOrder: "robin",
    };
  },
  methods: {
    createRoom: function() {
      console.log(this.roomName, this.musicSource, this.shuffleOrder);
    },
  }
}
</script>
