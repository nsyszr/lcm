<template>
  <el-container>
    <el-header height="40px" style="line-height: 40px;">
      <el-row type="flex" class="header" justify="space-between">
        <el-col :span="6" class="left">
          <span class="small-title">WebSocket Client</span>
        </el-col>
        <el-col :span="6" class="center"></el-col>
        <el-col :span="6" class="right"></el-col>
      </el-row>
    </el-header>
    <el-container style="border: 10px solid #e0e0e0;">
      <el-aside width="280px" style="border-right: 5px solid #e0e0e0;">
        <el-row>
          <el-col :span="24" style="padding: 20px 10px 10px 20px;">
            <el-button type="primary" size="mini" @click="connect">Connect</el-button>
            <el-button type="primary" size="mini" @click="disconnect">Disconnect</el-button>
            <el-button type="primary" size="mini" @click="sendHelloMessage">Send Hello (Registered)</el-button>
            <el-button
              type="primary"
              size="mini"
              @click="sendInvalidHelloMessage"
            >Send Hello (Unregistered)</el-button>
            <el-button type="primary" size="mini" @click="sendShitMessage">Send Shit</el-button>
            <el-button type="primary" size="mini" @click="sendEchoMessage">Send Echo</el-button>
          </el-col>
        </el-row>
      </el-aside>
    </el-container>
  </el-container>
</template>

<script>
var conn = null;

export default {
  name: "WebSocketClient",
  data() {
    return {};
  },
  methods: {
    connect() {
      this.$store.dispatch("deviceControlConnect");
    },
    disconnect() {
      this.$store.dispatch("deviceControlDisconnect");
    },
    sendHelloMessage() {
      this.$socket.send("HELLO:test");
    },
    sendInvalidHelloMessage() {
      this.$socket.send("HELLO:blaaa");
    },
    sendShitMessage() {
      this.$socket.send("SHIT");
    },
    sendEchoMessage() {
      this.$socket.send("echo this!");
    },
    refresh() {
      /*this.selected = [];
      this.$store
        .dispatch("managedDevices/fetchAllManagedDevices")
        .then(data => {
          this.managedDevices = data;
          if (data.length > 0) {
            this.setCurrent(this.managedDevices[0]);
          }
        });*/
    }
  },
  created() {
    /*this.refresh();*/
  }
};
</script>
