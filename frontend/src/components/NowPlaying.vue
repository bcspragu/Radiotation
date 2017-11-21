<template>
  <div class="now-playing">
    <img class="now-art" :src="image">
    <div class="metadata-holder">
        <div class="title metadata">{{Name}}</div>
        <div class="artist metadata">{{artist}}</div>
        <div class="album metadata">{{Album.Name}}</div>
    </div>
    <div class="veto" v-show="hasTrack">
      <button class="btn btn-lg">Veto <i class="icon icon-delete"></i>
      </button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'NowPlaying',
  props: ['Artists', 'Name', 'ID', 'Album', 'NoTrack'],
  computed: {
    artist () {
      var names = []
      for (const artist of this.Artists) {
        names.push(artist.Name)
      }
      return names.join(', ')
    },
    image () {
      var url = 'http://via.placeholder.com/150x150'
      if (!this.Album) {
        return url
      }
      if (this.Album.Images.length > 0) {
        url = this.Album.Images[0].URL
      }
      return url
    },
    hasTrack () {
      if (this.NoTrack) {
        return false
      }
      return true
    }
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

.veto {
  margin-right: 8px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
</style>
