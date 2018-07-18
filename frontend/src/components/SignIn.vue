<template>
  <div ref="signinBtn" class="button">Sign In</div>
</template>

<script>
export default {
  name: 'SignIn',
  data () {
    return {}
  },
  mounted () {
    window.gapi.load('auth2', () => {
      const auth2 = window.gapi.auth2.init({
        client_id: '990909845123-uhul9sqfjifu4f862phj18e4pr9lc5et.apps.googleusercontent.com'
      })
      auth2.attachClickHandler(this.$refs.signinBtn, {}, googleUser => {
        this.$emit('done', googleUser)
        // eslint-disable-next-line
      }, error => console.log(error))
    })
  }
}
</script>
