<template>
  <div class="results">
    <div class="input-group">
      <input 
        autocomplete="off"
        v-on:keyup.enter="search"
        type="text"
        v-model="query"
        name="search"
        class="form-input input-lg"
        placeholder="Search for Music">
      <button v-on:click="search" class="btn btn-lg input-group-btn"><i class="icon icon-search"></i></button>
    </div>
    <button v-on:click="goBack" class="btn btn-link btn-back btn-lg"><i class="icon icon-arrow-left back-icon"></i>Back to Queue</button>
    <div>
      <div v-show="noResults" class="no-results text-center">{{noResultsMsg}}</div>
      <div v-for="(track, index) in results" class="container" :key="track.Artist+track.Title+track.Image">
        <div class="columns col-gapless">
          <div class="column col-10"><track-item v-bind="track"/></div>
          <div class="column col-2 song-op" v-on:click="updateQueueStatus(track, index)">
            <button class="btn btn-link" :class="{disabled: track.InQueue}"><i class="icon" :class="addOrCheck(track.InQueue)"></i></button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import Track from '@/components/Track.vue'

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
    this.$emit('updateTitle', 'Loading results...')
    this.search()
  },
  computed: {
    noResults () {
      return this.results.length === 0
    }
  },
  methods: {
    addOrCheck (inQueue) {
      return {
        'icon-plus': !inQueue,
        'icon-check': inQueue
      }
    },
    updateQueueStatus (track, index) {
      if (track.InQueue) {
        return
      }
      this.addSong(track, index)
    },
    addSong (track, index) {
      var url = `/room/${this.roomID}/add`
      var data = {id: track.ID}
      this.$http.post(url, data, {emulateJSON: true}).then(response => {
        var data = response.body
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
      var url = `/room/${this.roomID}/search`
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
        this.$emit('updateTitle', 'Results for ' + this.query)
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
