<template>
  <v-app class="nsys-app">
    <app-toolbar/>
    <router-view class="nsys-content"/>
    <!-- <div id="nav">
      <router-link to="/">Home</router-link> |
      <router-link to="/about">About</router-link>
    </div>-->
  </v-app>
</template>

<script>
import AppToolbar from "@/components/AppToolbar.vue";
export default {
  name: "app",
  components: {
    AppToolbar
  }
};
</script>

<style scoped>
.nsys-app {
  background-color: #ffffff;
}
.nsys-content {
  padding-top: 64px;
}
</style>
