<template>
  <div class="track container">
    <div class="columns col-gapless">
      <div class="thumbnail column col-2"> <img :src="image">
      </div>
      <div class="metadata-holder column col-10">
        <div class="title metadata">{{Name}}</div>
        <div class="arist metadata">{{artist}}</div>
        <div class="album metadata">{{Album.Name}}</div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Track',
  props: ['Artists', 'Name', 'ID', 'Album'],
  computed: {
    image () {
      var url = 'http://via.placeholder.com/150x150'
      if (this.Album.Images.length > 0) {
        url = this.Album.Images[0].URL
      }
      return url
    },
    backgroundImage () {
      var url = 'http://via.placeholder.com/150x150'
      if (this.Album.Images.length > 0) {
        url = this.Album.Images[0].URL
      }
      return {
        'background-image': 'url("' + url + '")'
      }
    },
    artist () {
      var names = []
      for (const artist of this.Artists) {
        names.push(artist.Name)
      }
      return names.join(', ')
    }
  }
}
</script>

<style scoped>
.track {
  height: 80px;
  padding-top: 6px;
  padding-bottom: 6px;
}

.thumbnail {
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.thumbnail img {
  max-width: 100%;
}

div.metadata-holder.column {
  padding-left: 12px;
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
</style>
