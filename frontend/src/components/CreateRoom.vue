<template>
  <div>
    <h5 class="header text-center">Room Code '{{id}}' doesn't exist</h5>
    <h6 class="subheader text-center">Want to make a new room?</h6>
    <room-form></room-form>
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
.header {
  margin-top: 6px;
}

.subheader {
  margin-top: -8px;
}
</style>
