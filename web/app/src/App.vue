<template>
  <el-container id="app">
    <el-header style="background-color: #ececec;border-bottom: 1px solid #e0e0e0;">
      <el-row type="flex" class="header" justify="space-between">
        <el-col :span="6" class="left">
          <div>
            <img
              src="./assets/icom_logo.svg"
              class="logo"
              style="padding-top: 3px;background-color: #ececec;"
            >
            <span class="title" style="position: absolute; color: #909399;">Smart IoT Console</span>
          </div>
        </el-col>
        <el-col :span="6" class="center"></el-col>
        <el-col :span="6" class="right">
          <font-awesome-icon icon="user" style="color: #00a5db;"/>
        </el-col>
      </el-row>
    </el-header>
    <!-- <hello-world msg="Hello World!"></hello-world> -->
    <router-view></router-view>
  </el-container>
</template>

<script>
import HelloWorld from "./components/HelloWorld.vue";

export default {
  name: "app",
  components: {
    HelloWorld
  },
  data() {
    return {
      activeIndex: "1",
      activeIndex2: "1"
    };
  },
  methods: {
    handleSelect(key, keyPath) {
      console.log(key, keyPath);
    },
    handleOpen(key, keyPath) {
      console.log(key, keyPath);
    },
    handleClose(key, keyPath) {
      console.log(key, keyPath);
    }
  }
};
</script>

<style lang="scss">
@import url("https://fonts.googleapis.com/css?family=Titillium+Web");

html {
  overflow-y: auto;
  height: 100vh;
}

body {
  margin: 0;
  padding: 0;
  min-height: 100vh;
}

input,
textarea,
select,
button {
  font-family: "Titillium Web", Helvetica, Arial, sans-serif;
}

#app {
  font-family: "Titillium Web", Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #000;
  min-height: 100vh;
  margin: 0;
  padding: 0;
  font-size: 14px;
}

.logo {
  height: 54px;
  margin-left: -8px;
}

.el-header,
.el-footer {
  line-height: 60px;
}

.el-header {
  background-color: #fff;
  color: #000;
  padding: 0 0 !important;
  // border-bottom: 2px solid #e0e0e0;
  /*-webkit-box-shadow: 0px 2px 5px 0px rgba(38, 38, 38, 0.5);
  -moz-box-shadow: 0px 2px 5px 0px rgba(38, 38, 38, 0.5);
  box-shadow: 0px 2px 5px 0px rgba(38, 38, 38, 0.5);*/
}

.el-footer {
  background-color: #262626;
}

.el-aside {
  background-color: #fff;
  color: #262626;
  //background-color: #fff;
  //color: #000;
  //border-top: 1px solid #ececec;
}
.el-menu.nav-main {
  background-color: #fff;
  color: #262626;
  border-left: 0;
}
.el-menu-item-icon {
  margin-right: 9px;
  width: 24px;
  vertical-align: middle;
  text-align: center;
  font-size: 14px;
}

/*.el-menu-vertical-demo > .el-menu-item.is-active {
  color: #fff;
  font-weight: bold;
}

.el-menu-vertical-demo > .el-menu-item {
  color: #fff;
}*/

.el-main {
  background-color: #fff; //
  color: #262626;
  padding: 0px !important;
}

/*body > .el-container {
  margin-bottom: 40px;
}

.el-container:nth-child(5) .el-aside,
.el-container:nth-child(6) .el-aside {
  line-height: 260px;
}

.el-container:nth-child(7) .el-aside {
  line-height: 320px;
}*/

.el-row.header {
  padding: 0 20px;
}

.el-col.left {
  text-align: left;
}
.el-col.center {
  text-align: center;
}
.el-col.right {
  text-align: right;
}

.el-button {
  border-radius: 0 !important;
  font-weight: 600 !important;
}

.small-title {
  font-size: 16px;
  font-weight: 600;
}
.title {
  font-size: 18px;
  font-weight: 600;
}
</style>
