import Vue from "vue";
import store from "@/store";

// Initial state
const state = {
  socket: {
    isConnected: false,
    message: "",
    reconnectError: false
  }
};

const actions = {
  /*deviceControlSendMessage(message) {
    Vue.prototype.$socket.send(String(message));
  },*/
  deviceControlConnect() {
    Vue.prototype.$connect("ws://localhost:4001/devicecontrol/v1", {
      store: store
    });
  },
  deviceControlDisconnect() {
    Vue.prototype.$disconnect();
  }
};

// mutations
const mutations = {
  SOCKET_ONOPEN(state, event) {
    Vue.prototype.$socket = event.currentTarget;

    state.socket.isConnected = true;
  },
  SOCKET_ONCLOSE(state) {
    state.socket.isConnected = false;
  },
  SOCKET_ONERROR(state, event) {
    console.error(state, event);
  },
  // default handler called for all methods
  SOCKET_ONMESSAGE(state, message) {
    console.log(message);
    state.socket.message = message;
  },
  // mutations for reconnect methods
  SOCKET_RECONNECT(state, count) {
    console.info(state, count);
  },
  SOCKET_RECONNECT_ERROR(state) {
    state.socket.reconnectError = true;
  }
};

export default {
  state,
  // getters,
  actions,
  mutations
};
