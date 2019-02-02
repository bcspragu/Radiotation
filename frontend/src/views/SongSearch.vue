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
      <div v-for="(track, index) in results" class="container" :key="track.track.id">
        <div class="columns is-gapless is-mobile">
          <div class="column is-10"><Track v-bind="track.track"/></div>
          <div class="column is-2 song-op">
            <b-dropdown :disabled="track.inQueue">
                <button class="button is-link" :disabled="track.inQueue" slot="trigger">
                  <b-icon :icon="addOrCheck(track.inQueue)"></b-icon>
                </button>
                <b-dropdown-item v-if="!track.inQueue" v-on:click="addNext(track, index)">Add Next</b-dropdown-item>
                <b-dropdown-item v-if="!track.inQueue" v-on:click="addLast(track, index)">Add Last</b-dropdown-item>
            </b-dropdown>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';
import { TrackResult } from '@/data';
import Track from '@/components/Track.vue';

@Component({
  components: {
    Track,
  },
})
export default class SongSearch extends Vue {
  private noResultsMsg = 'Loading...';
  private query = '';
  private roomID = '';
  private results: TrackResult[] = [];

  private created(): void {
    if (this.$route.query.query instanceof Array) {
      this.query = this.$route.query.query[0];
    } else {
      this.query = this.$route.query.query;
    }
    this.roomID = this.$route.params.id;

    this.$emit('updateTitle', `Results for "${this.query}"`);
    this.search();
  }

  get noResults(): boolean {
    return this.results.length === 0;
  }

  private addOrCheck(inQueue: boolean): string {
    if (inQueue) {
      return 'check';
    }
    return 'plus';
  }

  private addNext(track: TrackResult, index: number): void {
    if (track.inQueue) {
      return;
    }
    this.addSong(track, index, 'addNext');
  }

  private addLast(track: TrackResult, index: number): void {
    if (track.inQueue) {
      return;
    }
    this.addSong(track, index, 'addLast');
  }

  private addSong(track: TrackResult, index: number, addPosition: string): void {
    const url = `room/${this.roomID}/${addPosition}`;
    const req = {id: track.track.id};
    this.$http.post(url, req).then((response) => {
      const data = response.data;
      if (data.Error) {
        this.$emit('ajaxErr', data);
        return;
      }
      this.results[index].inQueue = true;
    });
  }

  private goBack(): void {
    this.$router.back();
  }

  private search(): void {
    if (!this.query) {
      return;
    }
    const url = `room/${this.roomID}/search`;
    const req = {query: this.query};
    this.$http.get(url, {params: req}).then((response) => {
      const data = response.data;
      if (data.RoomNotFound) {
        this.$router.push({name: 'CreateRoom', params: {id: this.roomID}});
        return;
      }
      if (data.Error) {
        this.$emit('ajaxErr', data);
        return;
      }
      this.results = data;
      if (this.noResults) {
        this.noResultsMsg = 'No results found';
      }
    });
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
