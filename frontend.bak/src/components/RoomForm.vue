<template>
  <div>
    <b-field grouped>
      <b-field expanded label="Room Name">
        <b-input
          autocomplete="off"
          @keyup.native.enter="createRoom"
          type="text"
          v-model="roomName"
          name="room"
          class="form-input"
          placeholder="Room Name"></b-input>
      </b-field>
      <b-field label="Shuffle Order">
        <b-select v-model="shuffleOrder" name="shuffleOrder" class="form-select">
          <option value="robin">Round Robin</option>
          <option value="shuffle">Fair Random</option>
          <option value="random">True Random</option>
        </b-select>
      </b-field>
      <b-field class="align-button" label=".">
        <p class="control">
          <button v-on:click="createRoom" class="button is-primary">Create Room</button>
        </p>
      </b-field>
    </b-field>
  </div>
</template>

<script>
export default {
  data () {
    return {
      roomName: '',
      shuffleOrder: 'robin'
    }
  },
  methods: {
    createRoom () {
      var data = {
        roomName: this.roomName,
        shuffleOrder: this.shuffleOrder
      }
      this.$http.post('room', data, {emulateJSON: true}).then(response => {
        var data = response.body
        if (data.Error) {
          this.$emit('ajaxErr', data)
          return
        }
        this.$router.push({name: 'Room', params: {id: data.ID}})
      })
    }
  }
}
</script>

<style>
.align-button label {
  visibility: hidden;
}
</style>
