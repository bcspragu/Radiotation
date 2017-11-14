<template>
  <div class="container content main-content">
    <div class="text-center">
      <h1>Welcome to Radiotation!</h1>
    </div>
    <h2>Here's how it works</h2>
    <ol class="how-it-works">
      <li>Log in with you Google Account.</li>
      <li>Join an existing room with your friends or create a new one.</li>
      <li>Search for your favorite songs, and add them to your playlist.</li>
      <li>Open up the Radiotation app for Android and start playing it back.</li>
    </ol>
    <p>
      Radiotation will handle the rest, giving everyone equal playtime in the
      car (as long as everyone has added music!)
    </p>
    <!--{{ if .User }}-->
      <div class="columns">
        <div class="column col-6">
          <room-form></room-form>
        </div>
        <div class="column col-6">
          <ul>
            <!--{{ range .Rooms }}-->
              <!--<li>{{ .DisplayName }}</li>-->
            <!--{{ end }}-->
          </ul>
        </div>
      </div>
    <!--{{ else }}-->
      <div class="columns signin-holder">
        <div id="g-signin"></div>
      </div>
    <!--{{ end }}-->
  </div>
</template>

<script>
import RoomForm from './RoomForm.vue'

export default {
  components: {
    'room-form': RoomForm
  },
  data: function() {
    return {
      user: {name: 'brandon'}
    }
  }
}
</script>
