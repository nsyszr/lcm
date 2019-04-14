import { sleep, getCookie } from "@/utils/helper";

// initial state
const state = {
  isAuthenticated: false,
  userToken: localStorage.getItem("usertoken"),
  userSession: localStorage.getItem("usersession")
};

// getters
const getters = {
  isAuthenticated(state) {
    return state.isAuthenticated;
  },
  userToken(state) {
    return state.userToken;
  },
  userSession(state) {
    return state.userSession;
  }
};

// actions
const actions = {
  startSession({ commit, dispatch }) {
    dispatch("setAppLoading", true);

    // If development mode is active we start the session and do not parse
    // the session cookies, since they don't exsists. Authorization in dev-mode
    // is done thru HTTP headers. This mode works only if the backend is
    // started with the dev-mode flag.
    if (
      process.env.VUE_APP_DEV_MODE &&
      process.env.VUE_APP_DEV_MODE === "yes"
    ) {
      // eslint-disable-next-line no-unused-vars
      return new Promise((resolve, reject) => {
        var data = {
          userToken: "",
          userSession: JSON.stringify({
            firstName: "",
            lastName: "Dev-Mode User"
          })
        };

        commit("setSession", data);
        sleep(1000).then(() => {
          dispatch("setAppLoading", false);
        });
        resolve();
      });
    }

    return new Promise((resolve, reject) => {
      // Fetch the SESSID cookie, it contains the ID token header and payload
      const sessCookie = getCookie("SESSID");
      if (sessCookie == "") {
        // console.log("SESSID cookie not found.");
        dispatch("clearSession");
        reject("SESSID cookie not found.");
        return;
      }

      // Split the ID token
      const idTokenComponents = sessCookie.split(".");

      // Check if theres only a header and a payload component
      if (idTokenComponents.length != 2) {
        dispatch("clearSession");
        reject("Invalid SESSID cookie content.");
        return;
      }

      // console.log("Extract user info from token");
      try {
        // Encode the ID token
        const idTokenPayload = atob(idTokenComponents[1]);
        // Get JSON
        const t = JSON.parse(idTokenPayload);

        var data = {
          userToken: idTokenComponents[1],
          userSession: JSON.stringify({
            firstName: t.given_name,
            lastName: t.family_name
          })
        };

        commit("setSession", data);
        sleep(1000).then(() => {
          dispatch("setAppLoading", false);
        });
        resolve();
      } catch (e) {
        dispatch("clearSession");
        reject(e);
      }
    });
  },

  clearSession({ commit }) {
    commit("clearSession");
  },

  logout({ commit, dispatch }) {
    dispatch("setAppLoading", true);
    commit("clearSession");
    sleep(1000).then(() => {
      // In development mode we skip the redirect to the logout page
      if (
        process.env.VUE_APP_DEV_MODE &&
        process.env.VUE_APP_DEV_MODE === "yes"
      ) {
        location.replace("/");
        return;
      }

      location.replace(process.env.VUE_APP_OAUTH2_LOGOUT_URL);
    });
  }
};

// mutations
const mutations = {
  setSession(state, data) {
    state.isAuthenticated = true;
    localStorage.setItem("usertoken", data.userToken);
    localStorage.setItem("usersession", data.userSession);
  },
  clearSession(state) {
    state.isAuthenticated = false;
    localStorage.removeItem("usertoken");
    localStorage.removeItem("usersession");
  }
};

export default {
  state,
  getters,
  actions,
  mutations
};
