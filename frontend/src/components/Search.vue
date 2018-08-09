<template>
  <div class="results">
    <div class="columns is-gapless is-centered is-mobile">
      <div class="column is-10">
        <b-field>
          <b-input expanded placeholder="Search for Music"
            type="search"
            autocomplete="off"
            @keyup.native.enter="search"
            v-model="query"
            name="search"
            class="form-input input-lg">
          </b-input>
          <p class="control">
            <button v-on:click="search" class="button is-link"><b-icon icon="magnify"></b-icon></button>
          </p>
        </b-field>
      </div>
    </div>
    <button v-on:click="goBack" class="button is-link">
        <b-icon icon="arrow-left"></b-icon>
        <span>Back to Queue</span>
    </button>
    <div>
      <div v-show="noResults" class="no-results text-center">{{noResultsMsg}}</div>
      <div v-for="(track, index) in results" class="container" :key="track.ID">
        <div class="columns is-gapless is-mobile">
          <div class="column is-10"><track-item v-bind="track"/></div>
          <div class="column is-2 song-op">
            <b-dropdown :disabled="track.InQueue">
                <button class="button is-link" :disabled="track.InQueue" slot="trigger">
                  <b-icon :icon="addOrCheck(track.InQueue)"></b-icon>
                </button>
                <b-dropdown-item v-if="!track.InQueue" v-on:click="addNext(track, index)">Add Next</b-dropdown-item>
                <b-dropdown-item v-if="!track.InQueue" v-on:click="addLast(track, index)">Add Last</b-dropdown-item>
            </b-dropdown>
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
    this.search()
  },
  computed: {
    noResults () {
      return this.results.length === 0
    }
  },
  methods: {
    addOrCheck (inQueue) {
      if (inQueue) {
        return 'check'
      }
      return 'plus'
    },
    addNext (track, index) {
      if (track.InQueue) {
        return
      }
      this.addSong(track, index, "addNext")
    },
    addLast (track, index) {
      if (track.InQueue) {
        return
      }
      this.addSong(track, index, "addLast")
    },
    addSong (track, index, addPosition) {
      var url = `room/${this.roomID}/${addPosition}`
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
      this.$emit('updateTitle', {mod: 'pop' })
      this.$router.back()
    },
    search () {
      if (!this.query) {
        return
      }
      var url = `room/${this.roomID}/search`
      var data = {query: this.query}
      this.$http.get(url, {params: data, emulateJSON: true}).then(response => {
        var data = response.body
        if (data.RoomNotFound) {
          this.$router.push({name: 'CreateRoom', params: {id: this.roomID}})
          return
        }
        if (data.Error) {
          this.$emit('ajaxErr', data)
          return
        }
        this.results = data
        this.$emit('updateTitle', {mod: 'add', item: {text: 'Results for ' + this.query, to: '#' } })
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
