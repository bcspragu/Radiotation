<template>
  <div class="room-container">
    <div class="columns is-gapless is-centered is-mobile">
      <div class="column is-10">
        <b-field>
          <b-input expanded placeholder="Search for music..."
            type="search"
            autocomplete="off"
            @keyup.native.enter="goToSearch"
            v-model="query"
            name="search"
            class="form-input input-lg"
            icon="magnify">
          </b-input>
          <p class="control">
            <button class="button is-primary" @click="goToSearch">Search</button>
          </p>
        </b-field>
      </div>
    </div>
    <div class="queue">
      <div v-for="(track, index) in queue" class="container" :class="{played: track.Played, 'not-played': !track.Played}" :key="track.ID">
        <div class="columns is-gapless is-mobile">
          <div class="column is-10"><track-item v-bind="track"/></div>
          <div v-if="!track.Played" class="column is-2 song-op">
            <button v-on:click="removeSong(track, index)" class="button is-link"><b-icon icon="close"></b-icon></button>
          </div>
        </div>
      </div>
    </div>
    <div class="divider">Now Playing</div>
    <div class="now-playing">
      <now-playing :room-id="id" :track="nowPlaying"/>
    </div>
  </div>
</template>

<script>
import NowPlaying from '@/components/NowPlaying.vue'
import Search from '@/components/Search.vue'
import Track from '@/components/Track.vue'

export default {
  name: 'Room',
  data () {
    return {
      id: this.$route.params.id,
      room: {ID: '', DisplayName: ''},
      nowPlaying: {
        Name: 'Nothing Playing Yet',
        Artists: [{Name: ''}],
        Album: {
          Name: '',
          Images: [{URL: 'https://via.placeholder.com/150x150'}]
        },
        NoTrack: true
      },
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
    this.fetchRoom(true)
    this.connectWebSocket()
  },
  methods: {
    removeSong (track, index) {
      var url = `room/${this.id}/remove`
      var data = {index: index, id: track.ID}
      this.$http.post(url, data, {emulateJSON: true}).then(response => {
        var data = response.body;
        if (data.Error) {
          this.$emit('ajaxErr', data)
          return
        }
        this.queue.splice(index, 1)
      })
    },
    fetchRoom () {
      this.$http.get(`room/${this.id}`).then(response => {
        var data = response.body
        if (data.RoomNotFound) {
          this.$router.push({name: 'CreateRoom', params: {id: this.id}})
          return
        }
        if (data.Error) {
          this.$emit('ajaxErr', data)
          return
        }
        this.$emit('updateTitle', `Room '${data.Room.DisplayName}'`)
        this.room = data.Room
        this.queue = data.Queue
        if (!data.Track) {
          this.nowPlaying = {
            Name: 'Nothing Playing Yet',
            Artists: [{Name: ''}],
            Album: {
              Name: '',
              Images: [{URL: 'https://via.placeholder.com/150x150'}]
            },
            NoTrack: true
          }
        } else {
          this.nowPlaying = data.Track
        }
      })
    },
    goToSearch () {
      if (!this.query) {
        return
      }
      this.$router.push({name: 'Search', params: {roomID: this.id}, query: {query: this.query}})
    },
    connectWebSocket () {
      if (window['WebSocket']) {
        var loc = window.location
        var newURI = "ws:"
        if (loc.protocol === "https:") {
            newURI = "wss:"
        }
        newURI += "//" + loc.host
        newURI += loc.pathname + `api/ws/room/${this.id}`
        var conn = new WebSocket(newURI)
        conn.onclose = (evt) => {
          console.log(evt)
        }
        conn.onmessage = (evt) => {
          this.fetchRoom()
        }
      }
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
  overflow: auto;
}

.divider {
  margin: 0;
  font-size: 10px;
  height: 18px;
  line-height: 16px;
  background-color: #555;
  color: white;
  text-align: center;
}

.played {
  opacity: 0.2;
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
