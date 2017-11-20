<template>
  <div class="results">
    <h4><button v-on:click="goBack" class="btn btn-link btn-back"><i class="icon icon-arrow-left"></i></button>Results for "{{query}}"</h4>
    <div class="result-list">
      <div v-for="track in results" class="container" :key="track.Artist+track.Title+track.Image">
        <div class="columns col-gapless">
          <div class="column col-10"><track-item v-bind="track"></track-item></div>
          <div class="column col-2 add" v-on:click="addSong(track.ID)">
            <button class="btn btn-link"><i class="icon icon-plus"></i></button>
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
    addSong (id) {
      var url = '/room/' + this.roomID + '/add'
      var data = {id: id}
      this.$http.post(url, data, {emulateJSON: true}).then(response => {
        var data = JSON.parse(response.body)
        if (data.Error) {
          console.log(data)
          return
        }
        console.log(data)
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
          console.log(data)
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

.add {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
