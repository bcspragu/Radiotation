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

<script>
import RoomForm from '@/components/RoomForm.vue'

export default {
  name: 'CreateRoom',
  data () {
    return {
      id: this.$route.params.id
    }
  },
  created () {
    this.$emit('updateTitle', 'Create a Room')
  },
  components: {
    'room-form': RoomForm
  }
}
</script>

<style scoped>
</style>
