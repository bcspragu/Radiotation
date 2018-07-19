<template>
  <div v-if='errMsg == ""' class="now-playing">
    <img class="now-art" :src="image">
    <div class="metadata-holder">
        <div class="title metadata">{{track.Name}}</div>
        <div class="artist metadata">{{artist}}</div>
        <div class="album metadata">{{track.Album.Name}}</div>
    </div>
    <div class="veto" v-show="hasTrack">
      <button v-on:click="veto" class="btn btn-lg">Veto <i class="icon icon-delete"></i>
      </button>
    </div>
  </div>
  <div class="toast toast-primary text-center" v-else>
    <button v-on:click="clearErr" class="btn btn-clear float-right"></button>
    {{errMsg}}
  </div>
</template>

<script>
export default {
  name: 'NowPlaying',
  props: ['track', 'roomId'],
  data () {
    return {
      errMsg: ''
    }
  },
  computed: {
    artist () {
      var names = []
      for (const artist of this.track.Artists) {
        names.push(artist.Name)
      }
      return names.join(', ')
    },
    image () {
      var url = 'https://via.placeholder.com/150x150'
      if (!this.track.Album) {
        return url
      }
      if (this.track.Album.Images.length > 0) {
        url = this.track.Album.Images[0].URL
      }
      return url
    },
    hasTrack () {
      if (this.track.NoTrack) {
        return false
      }
      return true
    }
  },
  methods: {
    veto () {
      var url = `/room/${this.roomId}/veto`
      this.$http.post(url, {}, {emulateJSON: true}).then(response => {
        var data = response.body
        if (data.NotLoggedIn || data.RoomNotFound) {
          this.$emit('ajaxErr', data)
          return
        }
        if (data.Error) {
          this.errMsg = data.Message
        }
      })
    },
    clearErr () {
      this.errMsg = ''
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

.toast {
  height: 100%;
}
</style>
