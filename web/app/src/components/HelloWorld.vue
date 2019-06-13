<template>
  <el-container>
    <el-header height="40px" style="line-height: 40px;">
      <el-row type="flex" class="header" justify="space-between">
        <el-col :span="6" class="left">
          <span class="small-title" style="padding-left: 10px">
            Smart Devices
            <span style="padding-left: 5px; color: #909399;">3</span>
          </span>
        </el-col>
        <el-col :span="6" class="center"></el-col>
        <el-col :span="6" class="right"></el-col>
      </el-row>
    </el-header>
    <el-container style="border: 10px solid #e0e0e0;">
      <el-aside width="480px" style="border-right: 5px solid #e0e0e0;">
        <el-row>
          <el-col :span="24" style="padding: 20px 10px 10px 20px;">
            <el-button type="primary" size="mini">
              <font-awesome-icon icon="plus" style="color: #fff;margin-right: 5px;"/>Add device
            </el-button>
          </el-col>
        </el-row>
        <el-row>
          <el-col :span="24">
            <el-table
              ref="singleTable"
              :data="managedDevices"
              highlight-current-row
              @current-change="handleCurrentChange"
              style="width: 100%;"
              cell-style="border: 0; cursor:pointer;"
            >
              <el-table-column prop="connectionStatus" label width="50%">
                <template>
                  <font-awesome-icon
                    icon="exchange-alt"
                    style="color: #00ff00; margin-left: 10px;"
                  />
                </template>
              </el-table-column>
              <el-table-column prop="name" label="Name"></el-table-column>
              <el-table-column prop="location" label="Location"></el-table-column>
              <el-table-column prop="networkPrimaryIPv4Address" label="Address"></el-table-column>
            </el-table>
          </el-col>
        </el-row>
      </el-aside>
      <el-main style="padding: 10px !important;">
        <el-card class="box-card" shadow="never" v-if="currentRow">
          <div slot="header" class="clearfix">
            <font-awesome-icon icon="exchange-alt" style="color: #00ff00; margin-right: 5px;"/>
            <span class="small-title">{{currentRow.name}}</span>
            <el-dropdown style="float: right;" size="mini" split-button type="primary">
              Actions
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item>Action 1</el-dropdown-item>
                <el-dropdown-item>Action 2</el-dropdown-item>
                <el-dropdown-item>Action 3</el-dropdown-item>
                <el-dropdown-item>Action 4</el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </div>
          <el-row style="padding-bottom: 10px;">
            <el-col :span="24">
              <span style="font-weight: 600;">General</span>
            </el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Status</span>
            </el-col>
            <el-col :span="18">{{currentRow.connectionStatus}}</el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Device Type</span>
            </el-col>
            <el-col :span="18">{{currentRow.deviceType}}</el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Location</span>
            </el-col>
            <el-col :span="18">{{currentRow.location}}</el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Asset Tag</span>
            </el-col>
            <el-col :span="18">{{currentRow.assetTag}}</el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Group</span>
            </el-col>
            <el-col :span="18">{{currentRow.group}}</el-col>
          </el-row>

          <el-row style="padding-top: 10px; padding-bottom: 10px;">
            <el-col :span="24">
              <span style="font-weight: 600;">Hardware</span>
            </el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Model</span>
            </el-col>
            <el-col :span="18">{{currentRow.hardwareModel}}</el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Revision</span>
            </el-col>
            <el-col :span="18">{{currentRow.hardwareRevision}}</el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Serial Number</span>
            </el-col>
            <el-col :span="18">{{currentRow.hardwareSerialNumber}}</el-col>
          </el-row>

          <el-row style="padding-top: 10px; padding-bottom: 10px;">
            <el-col :span="24">
              <span style="font-weight: 600;">Firmware</span>
            </el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Version</span>
            </el-col>
            <el-col :span="18">{{currentRow.firmwareVersion}}</el-col>
          </el-row>

          <el-row style="padding-top: 10px; padding-bottom: 10px;">
            <el-col :span="24">
              <span style="font-weight: 600;">Network Settings</span>
            </el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Hostname</span>
            </el-col>
            <el-col :span="18">{{currentRow.networkHostname}}</el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Domainname</span>
            </el-col>
            <el-col :span="18">{{currentRow.networkDomainname}}</el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Primary IPv4 Address</span>
            </el-col>
            <el-col :span="18">{{currentRow.networkPrimaryIPv4Address}}</el-col>
          </el-row>

          <el-row style="padding-top: 10px; padding-bottom: 10px;">
            <el-col :span="24">
              <span style="font-weight: 600;">Availability</span>
            </el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Status</span>
            </el-col>
            <el-col :span="18">{{currentRow.availabilityStatus}}</el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Last Message</span>
            </el-col>
            <el-col :span="18">{{currentRow.availabilityLastMessageAt}}</el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Session Timeout</span>
            </el-col>
            <el-col :span="18">{{currentRow.availabilitySessionTimeout}} seconds</el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Ping Interval</span>
            </el-col>
            <el-col :span="18">{{currentRow.availabilityPingInterval}} seconds</el-col>
          </el-row>
          <el-row>
            <el-col :span="6">
              <span style="color: #a0a0a0">Ping Response Interval</span>
            </el-col>
            <el-col :span="18">{{currentRow.availabilityPongResponseInterval}} seconds</el-col>
          </el-row>
        </el-card>
      </el-main>
    </el-container>
  </el-container>
</template>

<script>
export default {
  name: "HelloWorld",
  props: {
    msg: String
  },
  data() {
    return {
      managedDevices: [],
      currentRow: null
    };
  },
  methods: {
    setCurrent(row) {
      this.$refs.singleTable.setCurrentRow(row);
    },
    handleCurrentChange(val) {
      this.currentRow = val;
    },
    refresh() {
      this.selected = [];
      this.$store
        .dispatch("managedDevices/fetchAllManagedDevices")
        .then(data => {
          this.managedDevices = data;
          if (data.length > 0) {
            this.setCurrent(this.managedDevices[0]);
          }
        });
    }
  },
  created() {
    this.refresh();
  }
};
</script>

<style lang="scss" scoped>
.el-card {
  border-radius: 0;
  border: 0;
}

.el-table--group::after,
.el-table--border::after,
.el-table::before {
  background: transparent;
}
</style>
