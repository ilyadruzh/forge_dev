<template>
  <div>
    <topmenu />
    <router-view/>
  </div>
</template>

<script>
const TopMenu = require('./components/TopMenu.vue');

module.exports = {
  components: {
    TopMenu,
  },
};
</script>


<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
