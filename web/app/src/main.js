import Vue from "vue";
import "./plugins/vue-cookies";
import "./plugins/vue-native-websocket";
import App from "./App.vue";
import router from "./router";
import store from "./store";
import "./plugins/element.js";

import { library } from "@fortawesome/fontawesome-svg-core";
import {
  faCoffee,
  faServer,
  faHome,
  faCog,
  faShippingFast,
  faUser,
  faPlus,
  faExchangeAlt
} from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";

library.add(
  faCoffee,
  faServer,
  faHome,
  faCog,
  faShippingFast,
  faUser,
  faPlus,
  faExchangeAlt
);

Vue.component("font-awesome-icon", FontAwesomeIcon);

Vue.config.productionTip = false;

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount("#app");
