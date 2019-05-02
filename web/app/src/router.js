import Vue from "vue";
import Router from "vue-router";
// import Home from "@/views/Home.vue";
// import ManagedDevices from "@/views/ManagedDevices.vue";
import HelloWorld from "@/components/HelloWorld.vue";
import WebSocketClient from "@/components/WebSocketClient.vue";
import store from "@/store";

Vue.use(Router);

const router = new Router({
  mode: "history",
  /*hash: false,
  history: true,
  base: process.env.BASE_URL,*/
  routes: [
    {
      path: "/",
      name: "hello-world",
      component: HelloWorld,
      meta: {
        requiresAuth: true
      }
    },
    {
      path: "/wsc",
      name: "websocket-client",
      component: WebSocketClient,
      meta: {
        requiresAuth: true
      }
    },
    {
      // catch all use case*/
      path: "*",
      redirect: "/"
    }
  ]
});

router.beforeEach((to, from, next) => {
  // TODO: This is maybe very critical
  if (to.matched.some(record => record.meta.requiresAuth)) {
    if (store.getters.isAuthenticated) {
      next();
    } else {
      store
        .dispatch("startSession")
        .then(() => {
          next();
        })
        // eslint-disable-next-line no-unused-vars
        .catch(err => {
          var url = "/oauth2/login";
          location.replace(url);
          next(false);
        });
    }
  } else {
    next();
  }
});

export default router;
