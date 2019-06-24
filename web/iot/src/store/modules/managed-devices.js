import ManagedDeviceApi from "@/services/ManagedDeviceApi";
import { SET_MANAGED_DEVICE } from "../mutation-types";
import { createManagedDevice } from "@/models/ManagedDevice";

// initial state
const state = {
  managedDevicesMap: new Map()
};

// getters
const getters = {
  getManagedDevices(state) {
    return Array.from(state.managedDevicesMap.values());
  }
};

// actions
const actions = {
  fetchAllManagedDevices({ commit }) {
    return new Promise((resolve, reject) => {
      ManagedDeviceApi.getManagedDevices()
        .then(r => r.data)
        .then(data => {
          var items = [];
          data.forEach(function(element) {
            const item = createManagedDevice(element);
            items.push(item);
            commit(SET_MANAGED_DEVICE, item);
          });
          // commit(SET_MANAGED_DEVICES, items);
          resolve(items);
        })
        .catch(err => {
          reject(err);
        });
    });
  },
  createNewManagedDevice({ commit }, payload) {
    return new Promise((resolve, reject) => {
      ManagedDeviceApi.createManagedDevice(payload)
        .then(r => r.data)
        .then(data => {
          const item = createManagedDevice(data);
          commit(SET_MANAGED_DEVICE, item);
          resolve(item);
        })
        .catch(err => {
          reject(err);
        });
    });
  }
};

// mutations
const mutations = {
  /*[SET_MANAGED_DEVICES](state, data) {
    data.forEach(function(element) {
      state.managedDevicesMap.set(element.id, element);
    });
  },*/
  [SET_MANAGED_DEVICE](state, data) {
    state.managedDevicesMap.set(data.id, data);
  }
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
};
