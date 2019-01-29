<template>
  <div class="room-container">
    <div class="columns is-mobile">
      <div class="column is-7-desktop is-8-mobile is-offset-1-desktop">
        <b-field>
          <b-input expanded placeholder="Search for music..."
            type="search"
            autocomplete="off"
            @keyup.native.enter="goToSearch"
            v-model="query"
            name="search"
            icon="magnify">
          </b-input>
          <p class="control">
            <button class="button is-primary" @click="goToSearch">Search</button>
          </p>
        </b-field>
      </div>
      <div class="column is-3-desktop is-4-mobile">
        <a class="button is-outlined is-fullwidth is-static" v-show="room.ID">Code: {{room.ID}}</a>
      </div>
    </div>
    <div class="queue">
      <div v-for="(track, index) in queue" class="container" :class="{played: track.Played, 'not-played': !track.Played}" :key="track.ID">
        <div class="columns is-gapless is-mobile">
          <div class="column is-10"><Track v-bind="track.Track"/></div>
          <div v-if="!track.Played" class="column is-2 song-op">
            <button v-on:click="removeSong(track, index)" class="button is-link"><b-icon icon="close"></b-icon></button>
          </div>
        </div>
      </div>
    </div>
    <div class="divider">Now Playing</div>
    <div class="now-playing">
      <NowPlaying :room-id="id" :track="nowPlaying"/>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';
import NowPlaying from '@/components/NowPlaying.vue';
import Track from '@/components/Track.vue';
import { Room as JRoom, QueueTrack, Track as JTrack } from '@/data';

@Component({
  components: {
    NowPlaying,
    Track,
  },
})
export default class Room extends Vue {
  private id = this.$route.params.id;
  private room: JRoom = {id: '', displayName: ''};
  private nowPlaying: JTrack | null = {
    name: 'Nothing Playing Yet',
    artists: [{name: ''}],
    album: {
      name: '',
      images: [{url: 'https://via.placeholder.com/150x150'}],
    },
  };
  private queue: QueueTrack[] = [];
  private query = '';

  private created(): void {
    this.fetchRoom();
    this.connectWebSocket();
  }

  private removeSong(track: QueueTrack, index: number): void {
    const url = `room/${this.id}/remove`;
    const req = {queueTrackID: track.id};
    this.$http.post(url, req).then((response) => {
      const data = response.data;
      if (data.Error) {
        this.$emit('ajaxErr', data);
        return;
      }
      this.queue.splice(index, 1);
    });
  }

  private fetchRoom(): void {
    this.$http.get(`room/${this.id}`).then((response) => {
      const data = response.data;
      if (data.RoomNotFound) {
        this.$router.push({name: 'createRoom', params: {id: this.id}});
        return;
      }
      if (data.Error) {
        this.$emit('ajaxErr', data);
        return;
      }
      this.$emit('updateTitle', `Room '${data.Room.DisplayName}'`);
      this.room = data.Room;
      this.queue = data.Queue;
      if (!data.Track) {
        this.nowPlaying = {
          name: 'Nothing Playing Yet',
          artists: [{name: ''}],
          album: {
            name: '',
            images: [{url: 'https://via.placeholder.com/150x150'}],
          },
        };
      } else {
        this.nowPlaying = data.Track;
      }
    });
  }

  private goToSearch(): void {
    if (!this.query) {
      return;
    }
    this.$router.push({name: 'songSearch', params: {roomID: this.id}, query: {query: this.query}});
  }

  private connectWebSocket(): void {
    const supportsWebSockets = 'WebSocket' in window || 'MozWebSocket' in window;
    if (supportsWebSockets) {
      const loc = window.location;
      let newURI = 'ws:';
      if (loc.protocol === 'https:') {
          newURI = 'wss:';
      }
      newURI += '//' + loc.host;
      newURI += loc.pathname + `api/ws/room/${this.id}`;
      const conn = new WebSocket(newURI);
      conn.onclose = (evt) => {
        console.log(evt);
      };
      conn.onmessage = (evt) => {
        this.fetchRoom();
      };
    }
  }
}
</script>
