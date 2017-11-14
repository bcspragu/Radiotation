<template>
  <div class="container">
    <form class="room-form" action="/rooms" method="POST">
      <div class="row">
        <div class="eight columns offset-by-two">
          <label for="room">Room Name</label>
          <input type="text" name="room" class="room-name u-full-width" placeholder="Room Name" value="">
        </div>
      </div>

      <div class="row">
        <div class="four columns offset-by-two">
          <label for="music_source">Music Source</label>
          <select name="music_source" class="u-full-width">
            <!--<option value="gmusic">Google Play Music</option>-->
            <option value="spotify">Spotify</option>
          </select>
        </div>
        <div class="four columns">
          <label for="shuffle_order">Shuffle Order</label>
          <select name="shuffle_order" class="u-full-width">
            <option value="robin">Round Robin</option>
            <option value="shuffle">Fair Random</option>
            <option value="random">True Random</option>
          </select>
        </div>
      </div>

      <div class="row">
        <button type="submit" class="button eight columns offset-by-two">Create Room</button>
      </div>
    </form>
  </div>
</template>
