<template>
  <div class="results">
    <h4><button v-on:click="goBack" class="btn btn-link btn-back"><i class="icon icon-arrow-left"></i></button>Results for "{{query}}"</h4>
    <div>
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
      roomID: this.$route.params.id,
      results: [],
      query: this.$route.query.query
    }
  },
  components: {
    'track-item': Track
  },
  created () {
    this.search()
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
        if (data.Error) {
          this.$emit('ajaxErr', data)
          return
        }
        this.results = data
      })
    }
  }
}
</script>

<style scoped>
.btn-back {
  margin: 6px;
  margin-bottom: 8px;
}

.song-op {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
