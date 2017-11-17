<template>
  <div class="flexbox-wrapper">
    <h1>Room {{id}}</h1>
    <div class="flexbox-row queue-holder">
      <div class="queue">
        <ol>
          <li v-for="track in tracks">{{track}}</li>
        </ol>
      </div>
    </div>
    <div class="flexbox-row flexbox-row-fill">
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
      <div class="results"></div>
    </div>
    <div class="flexbox-row flexbox-row-fixed now-playing">
      <!--{{ template "playing" .Tracks }}-->
    </div>
  </div>
</template>

<script>
export default {
  name: 'Room',
  data () {
    return {
      id: this.$route.params.id,
      results: [],
      tracks: [],
      query: ''
    }
  },
  methods: {
    fetchRoom () {
      this.$http.get('/rooms/' + this.id).then(response => {
        var data = JSON.parse(response.body)
        if (data.Error) {
          // Go to create page
          console.log(data)
          return
        }
        console.log(data)
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
        this.tracks = data
      })
    }
  }
}
</script>
