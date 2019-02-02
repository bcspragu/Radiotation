<template>
  <div v-if='errMsg == ""' class="now-playing">
    <img class="now-art" :src="image">
    <div class="metadata-holder">
        <div class="title metadata">{{track.name}}</div>
        <div class="artist metadata">{{artist}}</div>
        <div class="album metadata">{{track.album.name}}</div>
    </div>
  </div>
  <div class="toast toast-primary text-center" v-else>
    <button v-on:click="clearErr" class="btn btn-clear float-right"></button>
    {{errMsg}}
  </div>
</template>

<script lang="ts">
import { Component, Prop, Vue } from 'vue-property-decorator';
import { Track } from '@/data';

@Component
export default class NowPlaying extends Vue {
  @Prop({default: null}) private track!: Track | null;
  @Prop({default: ''}) private roomID!: string;

  private errMsg = '';

  get artist(): string {
    if (!this.track) {
      return '';
    }
    const names = [];
    for (const artist of this.track.artists) {
      names.push(artist.name);
    }
    return names.join(', ');
  }

  get image(): string {
    const defURL = 'https://via.placeholder.com/150x150';
    if (!this.track) {
      return defURL;
    }
    if (!this.track.album) {
      return defURL;
    }
    if (this.track.album.images.length > 0) {
      return this.track.album.images[0].url;
    }
    return defURL;
  }

  private clearErr(): void {
    this.errMsg = '';
  }
}
</script>

<style scoped>
.now-playing {
  display: flex;
}

.now-art {
  max-height: 100%;
}

.metadata-holder {
  overflow: hidden;
  margin-left: 6px;
  padding-right: 12px;

  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.metadata {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.title {
  font-size: 16px;
  font-weight: bold;
}

.artist {
  font-size: 12px;
}

.album {
  font-size: 10px;
}

.toast {
  height: 100%;
}
</style>
