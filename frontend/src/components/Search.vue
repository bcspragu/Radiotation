<template>
  <div class="results">
    <h4 class="header"><button v-on:click="goBack" class="btn btn-link btn-back"><i class="icon icon-arrow-left"></i></button>Results for "{{query}}"</h4>
    <div class="result-list">
      <track-item v-for="track in results" :key="track.Artist+track.Title+track.Image" v-bind="track"></track-item>
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
</style>
