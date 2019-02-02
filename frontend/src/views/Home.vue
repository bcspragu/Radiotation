<template>
  <div>
    <div class="columns">
      <div class="column is-6 is-offset-3 is-10-mobile is-offset-1-mobile">
        <h1 class="has-text-centered is-size-3">Search for Room</h1>
        <b-field grouped>
          <b-field expanded label="Room Code">
            <b-input
              autocomplete="off"
              @keyup.native.enter="search"
              type="text"
              v-model="query"
              name="room-code"
              placeholder="Room Code or Name"></b-input>
          </b-field>
          <b-field class="align-button" label=".">
            <p class="control">
              <button v-on:click="search" class="button is-primary">Search</button>
            </p>
          </b-field>
        </b-field>
        <h1 class="has-text-centered is-size-3">New Room</h1>
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
        <b-field grouped>
          <b-field label="Shuffle Order" expanded>
            <b-select v-model="shuffleOrder" name="shuffleOrder" class="form-select" expanded>
              <option value="robin">Round Robin</option>
              <option value="shuffle">Fair Random</option>
              <option value="random">True Random</option>
            </b-select>
          </b-field>
          <b-field class="align-button" label=".">
            <p class="control">
              <button v-on:click="createRoom" class="button is-primary">Create</button>
            </p>
          </b-field>
        </b-field>
      </div>
    </div>
    <hr>
    <div class="columns">
      <div class="column is-6 is-offset-3 is-10-mobile is-offset-1-mobile instructions">
        <h1 class="is-size-3 has-text-centered">Instructions</h1>
        <div class=columns>
          <div class="column is-10 is-offset-2">
            <ol class="is-size-4">
              <li>Join an existing room with your friends or create a new one.</li>
              <li>Search for your favorite songs, and add them to your playlist.</li>
              <li>Open up the Radiotation app for Android and start playing it back.</li>
            </ol>
          </div>
        </div>
        <hr>
        <div class=columns>
          <div class="column is-8 is-offset-2">
            <p class="is-size-5">
              Radiotation will handle the rest, giving everyone equal playtime in the
              car (as long as everyone has added music!)
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';

@Component
export default class Home extends Vue {
  // For searching.
  private query = '';

  // For creating.
  private roomName = '';
  private shuffleOrder = 'robin';

  private created(): void {
    this.$emit('updateTitle', 'Radiotation');
  }

  private search(): void {
    this.$http.get('search', {params: {query: this.query}}).then((response) => {
      const data = response.data;
      switch (data.type) {
        case 'results':
          this.$root.$data.searchResults = data.results;
          this.$router.push({name: 'RoomSearch', query: {query: this.query}});
          break;
        case 'room':
          this.$root.$data.roomInfo = data.roomInfo;
          this.$router.push({name: 'Room', params: {id: data.roomInfo.room.id}});
          break;
        default:
          console.log(`unknown response type ${data.type}`);
      }
    });
  }

  private createRoom(): void {
    const req = {
      roomName: this.roomName,
      shuffleOrder: this.shuffleOrder,
    };
    this.$http.post('room', req).then((response) => {
      const data = response.data;
      console.log(data);
      if (data.Error) {
        this.$emit('ajaxErr', data);
        return;
      }
      this.$router.push({name: 'Room', params: {id: data.ID}});
    });
  }
}
</script>

<style>
.align-button .label {
  visibility: hidden;
}

.columns {
	margin: 0;
}

@media screen and (min-width: 768px) {
	.columns {
	    margin-left: -0.75rem;
	    margin-right: -0.75rem;
	    margin-top: -0.75rem;
	}
}
</style>
