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

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';

interface Room {
  DisplayName: string;
  RoomCode: string;
  NumberUsers: number;
}

@Component
export default class RoomSearch extends Vue {
  private noResultsMsg = 'Loading...';
  private query = this.$route.query.query;
  private results: Room[] = [];

  private created(): void {
    this.$emit('updateTitle', `Results for room "${this.query}"`);
    this.searchRooms();
  }

  get noResults(): boolean {
    return this.results.length === 0;
  }

  private goBack(): void {
    this.$router.back();
  }

  private searchRooms(): void {
    const params = {query: this.query};
    this.$http.get('search', {params}).then((response) => {
      const data = response.data;
      this.results = data;
      if (this.results.length === 0) {
        this.noResultsMsg = 'No results found';
      }
    });
  }
}
</script>
