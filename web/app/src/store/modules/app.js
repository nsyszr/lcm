// initial state
const state = {
  appTitle: "Lifecycle Manager",
  appLoading: true,
  loading: false,
  error: undefined,
  fatalError: undefined
};

// getters
const getters = {
  appTitle(state) {
    return state.appTitle;
  },
  appLoading(state) {
    return state.appLoading;
  },
  loading(state) {
    return state.loading;
  },
  hasError(state) {
    return state.error !== undefined;
  },
  error(state) {
    return state.error;
  },
  hasFatalError(state) {
    return state.fatalError !== undefined;
  },
  fatalError(state) {
    return state.fatalError;
  }
};

// actions
const actions = {
  setAppTitle({ commit }, data) {
    commit("setAppTitle", data);
  },
  setAppLoading({ commit }, data) {
    commit("setAppLoading", data);
  },
  setLoading({ commit }, data) {
    commit("setLoading", data);
  },
  setError({ commit }, data) {
    var error = "An unknown error occurred. Please contact support.";
    if (data.hint !== undefined) {
      error = data.hint;
    } else if (data.description !== undefined) {
      error = data.description;
    } else if (data.error !== undefined) {
      error = data.error;
    } else if (data.message !== undefined) {
      error = data.message;
    }
    commit("setError", error);
  },
  clearError({ commit }) {
    commit("setError", undefined);
  },
  setFatalError({ commit }, data) {
    commit("setFatalError", data);
  },
  clearFatalError({ commit }) {
    commit("setFatalError", undefined);
  }
};

// mutations
const mutations = {
  setAppTitle(state, data) {
    state.appTitle = data;
  },
  setAppLoading(state, payload) {
    state.appLoading = payload;
  },
  setLoading(state, payload) {
    state.loading = payload;
  },
  setError(state, payload) {
    state.error = payload;
  },
  setFatalError(state, payload) {
    state.fatalError = payload;
  }
};

export default {
  state,
  getters,
  actions,
  mutations
};
