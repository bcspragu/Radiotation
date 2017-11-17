<template>
  <div class="flexbox-wrapper">
    <h1>{{room.DisplayName}}</h1>
    <div class="flexbox-row queue-holder">
      <div class="queue">
        <ol>
          <track-item v-for="track in queue.Tracks" :key="track.Artist+track.Title+track.Image" v-bind="track">{{track}}</track-item>
        </ol>
      </div>
    </div>
    <div class="container">
      <div class="search-form">
        <div class="row">
          <div class="seven columns offset-by-two">
            <input type="text" v-model="query" name="search" class="u-full-width" placeholder="Search for Music">
          </div>
          <div>
            <button v-on:click="search" class="two columns button button-primary search">Search</button>
          </div>
        </div>
      </div>
    </div>
    <div class="results">
    <track-item v-for="track in results" :key="track.Artist+track.Title+track.Image" v-bind="track"></track-item>
    </div>
    <div class="flexbox-row flexbox-row-fixed now-playing">
      <!--{{ template "playing" .Tracks }}-->
    </div>
  </div>
</template>

<script>
import Track from './Track.vue'

export default {
  name: 'Room',
  data () {
    return {
      id: this.$route.params.id,
      room: {ID: '', DisplayName: ''},
      nowPlaying: null,
      results: [],
      queue: {Tracks: []},
      query: ''
    }
  },
  components: {
    'track-item': Track
  },
  created () {
    this.fetchRoom()
  },
  methods: {
    fetchRoom () {
      this.$http.get('/room/' + this.id).then(response => {
        var data = JSON.parse(response.body)
        if (data.Error) {
          // Go to create page
          console.log(data)
          return
        }
        this.room = data.Room
        this.queue = data.Queue
        this.nowPlaying = data.Track
      })
    },
    search () {
      var url = '/room/' + this.id + '/search'
      var data = {query: this.query}
      this.$http.get(url, {params: data, emulateJSON: true}).then(response => {
        var data = JSON.parse(response.body)
        if (data.Error) {
          console.log(data)
          return
        }
        this.results = data
      })
    }
  }
}
</script>
