<template>
  <div class="room-container">
    <div class="columns">
      <div class="column is-7 is-offset-1">
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
      <div class="column is-3">
        <div class="room-id-button" @click="copyID">
        <a class="button is-outlined is-fullwidth is-static" v-show="room.id">Code: {{room.id}}</a>
        </div>
      </div>
    </div>
    <div class="queue">
      <div v-for="(track, index) in queue" class="container" :class="{played: track.played, 'not-played': !track.played}" :key="track.id">
        <div class="columns is-gapless is-mobile">
          <div class="column is-10"><Track v-bind="track.track"/></div>
          <div v-if="!track.played" class="column is-2 song-op">
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
import { Room as JRoom, QueueTrack, Track as JTrack, RoomInfo } from '@/data';

@Component({
  components: {
    NowPlaying,
    Track,
  },
})
export default class Room extends Vue {
  private id = '';
  private room: JRoom = {id: '', displayName: ''};
  private nowPlaying: JTrack | null = {
    id: '',
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
    this.id = this.$route.params.id;
    if (this.$root.$data.roomInfo) {
      const info = JSON.parse(JSON.stringify(this.$root.$data.roomInfo));
      this.setInfo(info);
      this.$root.$data.roomInfo = null;
    } else {
      this.fetchRoom();
    }
    this.connectWebSocket();
  }

  private setInfo(rmInfo: RoomInfo): void {
    this.room = rmInfo.room;
    this.queue = rmInfo.queue;
    if (!rmInfo.track) {
      this.nowPlaying = {
        id: '',
        name: 'Nothing Playing Yet',
        artists: [{name: ''}],
        album: {
          name: '',
          images: [{url: 'https://via.placeholder.com/150x150'}],
        },
      };
    } else {
      this.nowPlaying = rmInfo.track;
    }
    this.$emit('updateTitle', `Room '${rmInfo.room.displayName}'`);
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
        // TODO: Also add a note like, "Room ABCD" doesn't exist.
        this.$router.push({name: 'Home'});
        return;
      }
      if (data.Error) {
        this.$emit('ajaxErr', data);
        return;
      }
      this.setInfo(data);
    });
  }

  private goToSearch(): void {
    if (!this.query) {
      return;
    }
    this.$router.push({name: 'SongSearch', params: {roomID: this.id}, query: {query: this.query}});
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
      newURI += `/api/ws/room/${this.id}`;
      const conn = new WebSocket(newURI);
      conn.onclose = (evt) => {
        console.log('closed');
        console.log(evt);
      };
      conn.onmessage = (evt) => {
        console.log('msg');
        this.fetchRoom();
      };
    }
  }

  private copyID(): void {
    // Create a <textarea> element
    const el = document.createElement('textarea');
    // Set its value to the string that you want copied
    el.value = this.id;
    // Make it readonly to be tamper-proof
    el.setAttribute('readonly', '');
    el.style.position = 'absolute';
    // Move outside the screen to make it invisible
    el.style.left = '-9999px';
    // Append the <textarea> element to the HTML document
    document.body.appendChild(el);
    let selected!: boolean | Range;
    if (document && document.getSelection()) {
      // Check if there is any content selected previously.
      selected = document.getSelection()!.rangeCount > 0
          ? document!.getSelection()!.getRangeAt(0) // Store selection if found
          : false;                                // Mark as false to know no selection existed before
    }

    // Select the <textarea> content
    el.select();
    // Copy - only works as a result of a user action (e.g. click events)
    document.execCommand('copy');
    // Remove the <textarea> element
    document.body.removeChild(el);

    // If a selection existed before copying
    if (selected instanceof Range) {
      // Unselect everything on the HTML document
      document.getSelection()!.removeAllRanges();
      // Restore the original selection
      document.getSelection()!.addRange(selected);
    }
  }
}
</script>

<style scoped>
.room-container {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.queue {
  flex: 1;
  overflow: auto;
}

.divider {
  margin: 0;
  font-size: 10px;
  height: 18px;
  line-height: 16px;
  background-color: #555;
  color: white;
  text-align: center;
}

.played {
  opacity: 0.2;
}

.song-op {
  display: flex;
  align-items: center;
  justify-content: center;
}

.now-playing {
  background-color: #F8F9FA;
  height: 75px;
}

.room-id-button:hover {
  cursor: pointer;
}
</style>
