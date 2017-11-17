<template>
  <div class="track columns">
    <div class="img">
      <img class="img-responsive column" :src="image"/>
    </div>
    <div class="column">
      <div class="title">{{Name}}</div>
      <div class="arist">{{artist}}</div>
      <div class="album">{{Album.Name}}</div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Track',
  props: ['Artists', 'Name', 'ID', 'Album'],
  computed: {
    image () {
      if (this.Album.Images.length === 0) {
        return 'http://via.placeholder.com/150x150'
      }
      return this.Album.Images[0].URL
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
  height: 100px;
}

.img {
  width: 100px;
}

.track img {
  display: block;
  width: 100%;
  height: auto;
}

.title {
  font-size: 20px;
  font-weight: bold;
}

.artist {
  font-size: 16px;
}

.album {
  font-size: 14px;
}
</style>
