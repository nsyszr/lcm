import axios from "axios";
import store from "@/store";

export default loading => {
  // Create to axios instances. One with loading state, another without loading
  // state. Last on is used for API calls like 'check-alias'. This API is
  // called during validation and we UI would flicker or it never leaves
  // the loading state.

  var a = axios.create({
    baseURL: process.env.VUE_APP_API_BASE_URL,
    // withCredentials: true,
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json"
    }
  });

  if (process.env.VUE_APP_DEV_MODE && process.env.VUE_APP_DEV_MODE === "yes") {
    a.interceptors.request.use(
      config => {
        store.dispatch("setLoading", true);
        config.headers["x-insys-debug-username"] =
          process.env.VUE_APP_DEV_MODE_DEFAULT_USERNAME;
        return config;
      },
      err => {
        store.dispatch("setLoading", false);
        return Promise.reject(err);
      }
    );
  } else {
    // The interceptor for production mode. The interceptor controls the loading
    // state of the application.
    a.interceptors.request.use(
      config => {
        store.dispatch("setLoading", true);
        return config;
      },
      err => {
        store.dispatch("setLoading", false);
        return Promise.reject(err);
      }
    );
  }

  a.interceptors.response.use(
    response => {
      store.dispatch("setLoading", false);
      return response;
      // new Promise(resolve => setTimeout(() => resolve(response), 2000));
    },
    error => {
      store.dispatch("setLoading", false);

      // On network error set fatal error
      if (!error.response) {
        store.dispatch("setFatalError", "Network error.");
      }
      // On 401 immediately invalide session with logout
      else if (error.response.status === 401) {
        store.dispatch("logout");
      } else if (error.response.status >= 500) {
        store.dispatch("setFatalError", error.response.data);
      } else if (error.response.status >= 400) {
        store.dispatch("setError", error.response.data);
      }
      return Promise.reject(error.response);
    }
  );

  // Axios instance without loading state modification
  var aNonLoading = axios.create({
    baseURL: process.env.VUE_APP_API_BASE_URL,
    // withCredentials: true,
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json"
    }
  });

  if (process.env.VUE_APP_DEV_MODE && process.env.VUE_APP_DEV_MODE === "yes") {
    aNonLoading.interceptors.request.use(
      config => {
        config.headers["X-INSYS-Debug-Username"] =
          process.env.VUE_APP_DEV_MODE_DEFAULT_USERNAME;
        return config;
      },
      err => {
        return Promise.reject(err);
      }
    );
  } else {
    // The interceptor for production mode. The interceptor controls the loading
    // state of the application.
    aNonLoading.interceptors.request.use(
      config => {
        return config;
      },
      err => {
        return Promise.reject(err);
      }
    );
  }

  aNonLoading.interceptors.response.use(
    response => {
      return response;
      // new Promise(resolve => setTimeout(() => resolve(response), 2000));
    },
    error => {
      // On network error set fatal error
      if (!error.response) {
        store.dispatch("setFatalError", "Network error.");
      }
      // On 401 immediately invalide session with logout
      else if (error.response.status === 401) {
        store.dispatch("logout");
      } else if (error.response.status >= 500) {
        store.dispatch("setFatalError", error.response.data);
      } else if (error.response.status >= 400) {
        store.dispatch("setError", error.response.data);
      }
      return Promise.reject(error.response);
    }
  );

  if (loading) {
    return a;
  }

  return aNonLoading;
};
