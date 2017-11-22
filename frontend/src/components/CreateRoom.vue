<template>
  <div>
    <h5 class="header text-center">Room '{{id}}' doesn't exist</h5>
    <h6 class="subheader text-center">You can create it here.</h6>
    <room-form :default-name="id"></room-form>
  </div>
</template>

<script>
import RoomForm from './RoomForm.vue'

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
