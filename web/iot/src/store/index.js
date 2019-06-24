import Vue from "vue";
import Vuex from "vuex";
import managedDevices from "./modules/managed-devices";
import createLogger from "@/utils/logger";

Vue.use(Vuex);

const debug = process.env.NODE_ENV !== "production";

export default new Vuex.Store({
  modules: {
    managedDevices
  },
  strict: debug,
  plugins: debug ? [createLogger()] : []
});
