<template>
  <div class="results">
    <div class="columns">
      <h1 class="is-size-2 column is-10 is-offset-1">Results for "{{ query }}"</h1>
    </div>
    <div v-show="noResults" class="no-results text-center">{{noResultsMsg}}</div>
    <div v-for="room in results" :key="room.RoomCode">
      <div class="columns is-centered">
        <router-link :to="{ name: 'Room', params: { id: room.RoomCode } }" class="column is-9 is-size-3">
          {{room.DisplayName}}
        </router-link>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'RoomList',
  data () {
    return {
      noResultsMsg: 'Loading...',
      query: this.$route.query.query,
      results: [],
    }
  },
  created () {
    this.searchRooms()
  },
  computed: {
    noResults () {
      return this.results.length === 0
    }
  },
  methods: {
    goBack () {
      this.$router.back()
    },
    searchRooms () {
      var data = {query: this.query}
      this.$http.get('search', {params: data, emulateJSON: true}).then(response => {
        var data = response.body
        this.results = data
        if (this.results.length === 0) {
          this.noResultsMsg = 'No results found'
        }
      })
    }
  }
}
</script>

<style scoped>
</style>
