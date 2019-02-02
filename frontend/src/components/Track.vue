<template>
  <div class="track">
    <div class="columns is-gapless is-mobile">
      <div class="thumbnail column is-2">
        <img :src="image">
      </div>
      <div class="column is-10">
        <div class="metadata-holder">
          <div class="title metadata">{{name}}</div>
          <div class="artist metadata">{{artist}}</div>
          <div class="album metadata">{{album.name}}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Prop, Vue } from 'vue-property-decorator';
import { Artist, Album } from '@/data';

@Component
export default class Track extends Vue {
  @Prop({default: () => [] }) private artists!: Artist[];
  @Prop({default: ''}) private name!: string;
  @Prop({default: null}) private album!: Album | null;

  get artist(): string {
    const names: string[] = [];
    for (const artist of this.artists) {
      names.push(artist.name);
    }
    return names.join(', ');
  }

  get image(): string {
    let url = 'https://via.placeholder.com/150x150';
    if (!this.album) {
      return url;
    }
    if (this.album.images.length > 0) {
      url = this.album.images[0].url;
    }
    return url;
  }
}
</script>

<style scoped>
.track {
  padding-top: 6px;
  padding-bottom: 6px;
}

.thumbnail {
  padding: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.thumbnail img {
  max-width: 100%;
}

.metadata-holder {
  padding-left: 12px;
}

.metadata {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 0;
  margin: 0;
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
</style>
