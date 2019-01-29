<template>
  <div>
    <h1 class="is-size-2 has-text-centered">Room Code '{{id}}' doesn't exist</h1>
    <h2 class="is-size-3 has-text-centered">Want to make a new room?</h2>
    <div class="columns is-centered">
      <div class="column is-6 is-10-mobile is-offset-1-mobile">
        <room-form></room-form>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';
import RoomForm from '@/components/RoomForm.vue';

@Component({
  components: {
    RoomForm,
  },
})
export default class CreateRoom extends Vue {
  private id: string = this.$route.params.id;

  private created(): void {
    this.$emit('updateTitle', 'Create a Room');
  }
}
</script>
