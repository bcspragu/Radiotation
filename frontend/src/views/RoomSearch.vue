<template>
  <div class="results">
    <div class="columns">
      <h1 class="is-size-2 column is-10 is-offset-1">Results for "{{ query }}"</h1>
    </div>
    <div v-show="noResults" class="no-results text-center">{{noResultsMsg}}</div>
    <div v-for="room in results" :key="room.roomCode">
      <div class="columns is-centered">
        <router-link :to="{ name: 'Room', params: { id: room.roomCode } }" class="column is-9 is-size-3">
          {{room.displayName}} [{{room.numberUsers}} user(s)]
        </router-link>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';

interface Room {
  displayName: string;
  roomCode: string;
  numberUsers: number;
}

@Component
export default class RoomSearch extends Vue {
  private noResultsMsg = 'Loading...';
  private query = '';
  private results: Room[] = [];

  private created(): void {
    if (this.$route.query.query instanceof Array) {
      this.query = this.$route.query.query[0];
    } else {
      this.query = this.$route.query.query;
    }
    this.$emit('updateTitle', `Results for room "${this.query}"`);
    if (this.$root.$data.results) {
      const results = JSON.parse(JSON.stringify(this.$root.$data.searchResults));
      this.results = results;
      this.$root.$data.searchResults = [];
    } else {
      this.searchRooms();
    }
  }

  get noResults(): boolean {
    return this.results.length === 0;
  }

  private goBack(): void {
    this.$router.back();
  }

  private searchRooms(): void {
    this.$http.get('search', {params: {query: this.query}}).then((response) => {
      const data = response.data;
      switch (data.type) {
        case 'results':
          this.results = data.results;
          break;
        default:
          console.log(`expected response type results, can't use ${data.type} here`);
      }
    });
  }
}
</script>
