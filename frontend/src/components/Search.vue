<template>
  <div class="results">
    <button v-on:click="goBack" class="btn btn-link btn-back btn-lg"><i class="icon icon-arrow-left back-icon"></i>Back to Queue</button>
    <div>
      <div v-show="noResults" class="no-results text-center">{{noResultsMsg}}</div>
      <div v-for="(track, index) in results" class="container" :key="track.Artist+track.Title+track.Image">
        <div class="columns col-gapless">
          <div class="column col-10"><track-item v-bind="track"/></div>
          <div class="column col-2 song-op" v-on:click="updateQueueStatus(track, index)">
            <button class="btn btn-link"><i class="icon" :class="addOrRemove(track.InQueue)"></i></button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import Track from './Track.vue'

export default {
  name: 'Room',
  data () {
    return {
      noResultsMsg: 'Loading...',
      query: this.$route.query.query,
      results: [],
      roomID: this.$route.params.id
    }
  },
  components: {
    'track-item': Track
  },
  created () {
    this.$emit('updateTitle', 'Results for ' + this.query)
    this.search()
  },
  computed: {
    noResults () {
      return this.results.length === 0
    }
  },
  methods: {
    addOrRemove (inQueue) {
      return {
        'icon-plus': !inQueue,
        'icon-cross': inQueue
      }
    },
    updateQueueStatus (track, index) {
      if (track.InQueue) {
        this.removeSong(track, index)
      } else {
        this.addSong(track, index)
      }
    },
    removeSong (track, index) {
      var url = '/room/' + this.roomID + '/remove'
      var data = {index: track.Index, id: track.ID}
      this.$http.post(url, data, {emulateJSON: true}).then(response => {
        var data = JSON.parse(response.body)
        if (data.Error) {
          this.$emit('ajaxErr', data)
          return
        }
        this.results[index].InQueue = false
      })
    },
    addSong (track, index) {
      var url = '/room/' + this.roomID + '/add'
      var data = {id: track.ID}
      this.$http.post(url, data, {emulateJSON: true}).then(response => {
        var data = JSON.parse(response.body)
        if (data.Error) {
          this.$emit('ajaxErr', data)
          return
        }
        this.results[index].InQueue = true
      })
    },
    goBack () {
      this.$router.back()
    },
    search () {
      if (!this.query) {
        return
      }
      var url = '/room/' + this.roomID + '/search'
      var data = {query: this.query}
      this.$http.get(url, {params: data, emulateJSON: true}).then(response => {
        var data = JSON.parse(response.body)
        if (data.RoomNotFound) {
          this.$router.push({name: 'CreateRoom', params: {id: this.roomID}})
          return
        }
        if (data.Error) {
          this.$emit('ajaxErr', data)
          return
        }
        this.results = data
        if (this.noResults) {
          this.noResultsMsg = 'No results found'
        }
      })
    }
  }
}
</script>

<style scoped>
.back-icon {
  margin-right: 6px;
}

.no-results {
  margin-top: 12px;
  font-size: 24px;
}

.song-op {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
