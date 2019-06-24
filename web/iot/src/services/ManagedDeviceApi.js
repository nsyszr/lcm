import Api from "@/services/Api";

export default {
  getManagedDevices() {
    return Api(true).get("/managed-devices");
  },
  createManagedDevice(payload) {
    return Api(true).post("/managed-devices", payload);
  }
};
