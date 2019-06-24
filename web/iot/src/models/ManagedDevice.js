export class ManagedDevice {
  constructor({
    id = ``,
    deviceType = `UNDEFINED`,
    name = ``,
    location = ``,
    assetTag = ``,
    group = `UNASSIGNED`,
    hardwareModel = ``,
    hardwareRevision = ``,
    hardwareSerialNumber = ``,
    firmwareVersion = ``,
    networkHostname = ``,
    networkDomainname = ``,
    networkPrimaryIPv4Address = ``,
    availabilitySessionTimeout = 0,
    availabilityPingInterval = 0,
    availabilityPongResponseInterval = 0,
    availabilityLastMessageAt = null,
    availabilityStatus = `UNKNOWN`,
    connectionStatus = `UNKNOWN`
  } = {}) {
    this.id = id;
    this.deviceType = deviceType;
    this.name = name;
    this.location = location;
    this.assetTag = assetTag;
    this.group = group;
    this.hardwareModel = hardwareModel;
    this.hardwareRevision = hardwareRevision;
    this.hardwareSerialNumber = hardwareSerialNumber;
    this.firmwareVersion = firmwareVersion;
    this.networkHostname = networkHostname;
    this.networkDomainname = networkDomainname;
    this.networkPrimaryIPv4Address = networkPrimaryIPv4Address;
    this.availabilitySessionTimeout = availabilitySessionTimeout;
    this.availabilityPingInterval = availabilityPingInterval;
    this.availabilityPongResponseInterval = availabilityPongResponseInterval;
    this.availabilityLastMessageAt = availabilityLastMessageAt;
    this.availabilityStatus = availabilityStatus;
    this.connectionStatus = connectionStatus;
  }
}

export function createManagedDevice(data) {
  return Object.freeze(new ManagedDevice(data));
}
