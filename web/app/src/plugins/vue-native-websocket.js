import Vue from "vue";
import VueNativeSock from "vue-native-websocket";

Vue.use(VueNativeSock, "ws://localhost:4001/devicecontrol/v1", {
  connectManually: true
});
