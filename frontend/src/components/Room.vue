<template>
  <div class="room-container">
    <div class="input-group">
      <input type="text" v-model="query" name="search" class="form-input input-lg" placeholder="Search for Music">
      <button v-on:click="goToSearch" class="btn btn-lg input-group-btn"><i class="icon icon-search"></i></button>
    </div>
    <div class="queue">
      <div class="divider">Your Queue</div>
      <div class="queue">
        <ol>
          <track-item 
            v-for="track in queue.Tracks" 
            v-bind="track"
            v-on:click="remove"
            :key="track.Artist+track.Title+track.Image">{{track}}</track-item>
        </ol>
      </div>
    </div>
    <div>
      <now-playing/>
    </div>
  </div>
</template>

<script>
import NowPlaying from './NowPlaying.vue'
import Search from './Search.vue'
import Track from './Track.vue'

export default {
  name: 'Room',
  data () {
    return {
      id: this.$route.params.id,
      room: {ID: '', DisplayName: ''},
      nowPlaying: null,
      queue: {Tracks: []},
      query: ''
    }
  },
  components: {
    'now-playing': NowPlaying,
    'search': Search,
    'track-item': Track
  },
  created () {
    this.fetchRoom()
  },
  methods: {
    remove () {

    },
    fetchRoom () {
      this.$http.get('/room/' + this.id).then(response => {
        var data = JSON.parse(response.body)
        if (data.Error) {
          this.$router.push({name: 'CreateRoom', params: {id: this.id}})
          return
        }
        this.$emit('updateTitle', data.Room.DisplayName)
        this.room = data.Room
        this.queue = data.Queue
        this.nowPlaying = data.Track
      })
    },
    goToSearch () {
      this.$router.push({name: 'Search', params: {roomID: this.id}, query: {query: this.query}})
    }
  }
}
</script>

<style scoped>
.room-container {
  display: flex;
  flex-direction: column;
}

.queue {
  flex: 9;
}

.divider {
  margin: 0;
  font-size: 20px;
  height: 24px;
  line-height: 22px;
  background: #F8F9FA;
  border-top: 2px solid #F0F1F2;
  border-bottom: 2px solid #F0F1F2;
  font-color: white;
  text-align: center;
}
</style>
