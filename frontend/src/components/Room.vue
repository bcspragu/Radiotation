<template>
  <div class="room-container">
    <div class="input-group">
      <input v-on:keyup.enter="goToSearch" type="text" v-model="query" name="search" class="form-input input-lg" placeholder="Search for Music">
      <button v-on:click="goToSearch" class="btn btn-lg input-group-btn"><i class="icon icon-search"></i></button>
    </div>
    <div class="queue">
      <div class="divider">Your Queue</div>
      <div class="queue">
        <div v-for="(track, index) in queue" class="container" :key="track.Artist+track.Title+track.Image">
          <div class="columns col-gapless">
            <div class="column col-10"><track-item v-bind="track"/></div>
            <div class="column col-2 song-op" v-on:click="removeSong(track, index)">
              <button class="btn btn-link"><i class="icon icon-cross"></i></button>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="now-playing">
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
      queue: [],
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
    removeSong (track, index) {
      var url = '/room/' + this.id + '/remove'
      var data = {index: index, id: track.ID}
      this.$http.post(url, data, {emulateJSON: true}).then(response => {
        var data = JSON.parse(response.body)
        if (data.Error) {
          this.$emit('ajaxErr', data)
          return
        }
        this.queue.splice(index, 1)
      })
    },
    fetchRoom () {
      this.$http.get('/room/' + this.id).then(response => {
        var data = JSON.parse(response.body)
        if (data.NotLoggedIn) {
          this.$emit('ajaxErr', data)
          return
        }
        if (data.Error) {
          this.$router.push({name: 'CreateRoom', params: {id: this.id}})
          return
        }
        this.$emit('updateTitle', 'Room ' + data.Room.DisplayName)
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
  flex: 1;
  display: flex;
  flex-direction: column;
}

.queue {
  flex: 1;
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

.song-op {
  display: flex;
  align-items: center;
  justify-content: center;
}

.now-playing {
  background-color: #F8F9FA;
  height: 75px;
}
</style>
