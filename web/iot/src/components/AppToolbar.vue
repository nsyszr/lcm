<template>
  <div id="nav">
    <v-toolbar fixed class="nsys-toolbar">
      <!-- <v-toolbar-side-icon @click.stop="drawer = !drawer"></v-toolbar-side-icon> -->
      <!-- <v-toolbar-title @click.stop="drawer = !drawer" class="nsys-toolbar-icon">
        <img src="../assets/icom_logo.svg" alt>
      </v-toolbar-title>-->
      <v-menu v-model="menu" :close-on-content-click="false" absolute>
        <template v-slot:activator="{ on }">
          <v-toolbar-title v-on="on" class="black--text nsys-toolbar-title">
            Smart IoT Console
            <v-icon
              style="padding-left: 16px; padding-right: 8px;"
              color="primary"
            >keyboard_arrow_down</v-icon>
          </v-toolbar-title>
        </template>

        <v-card class="elevation-0">
          <v-card-title class="nsys-vcard-title" primary-title>
            <h3 class="nsys-menu-title mb-0">
              Smart IoT Console
              <v-icon
                style="padding-left: 16px; padding-right: 8px;"
                color="primary"
              >keyboard_arrow_down</v-icon>
            </h3>
          </v-card-title>
          <v-divider style="margin-left: 16px; margin-right: 16px;"/>
          <v-container style="padding: 8px;">
            <v-layout align-start justify-start row fill-height>
              <v-flex class="pr-4">
                <v-list dense subheader>
                  <v-subheader>General</v-subheader>
                  <v-list-tile @click="true">
                    <v-list-tile-content>
                      <v-list-tile-title>Profile</v-list-tile-title>
                    </v-list-tile-content>
                  </v-list-tile>
                  <v-list-tile @click="true">
                    <v-list-tile-content>
                      <v-list-tile-title>Organisation</v-list-tile-title>
                    </v-list-tile-content>
                  </v-list-tile>
                  <v-list-tile @click="true">
                    <v-list-tile-content>
                      <v-list-tile-title>Users</v-list-tile-title>
                    </v-list-tile-content>
                  </v-list-tile>
                  <v-list-tile @click="true">
                    <v-list-tile-content>
                      <v-list-tile-title>Subscriptions</v-list-tile-title>
                    </v-list-tile-content>
                  </v-list-tile>
                </v-list>
              </v-flex>
              <v-flex class="pr-4">
                <v-list dense subheader>
                  <v-subheader>Smart Devices</v-subheader>
                  <v-list-tile @click="true">
                    <v-list-tile-content>
                      <v-list-tile-title>Inventory</v-list-tile-title>
                    </v-list-tile-content>
                  </v-list-tile>
                  <v-list-tile @click="true">
                    <v-list-tile-content>
                      <v-list-tile-title>Update Jobs</v-list-tile-title>
                    </v-list-tile-content>
                  </v-list-tile>
                </v-list>
              </v-flex>
              <v-flex class="pr-4">
                <v-list dense subheader>
                  <v-subheader>Connectivity</v-subheader>
                  <v-list-tile @click="true">
                    <v-list-tile-content>
                      <v-list-tile-title>VPN Client Devices</v-list-tile-title>
                    </v-list-tile-content>
                  </v-list-tile>
                  <v-list-tile @click="true">
                    <v-list-tile-content>
                      <v-list-tile-title>Web-Proxies</v-list-tile-title>
                    </v-list-tile-content>
                  </v-list-tile>
                </v-list>
              </v-flex>
            </v-layout>
          </v-container>
        </v-card>
      </v-menu>
      <v-spacer></v-spacer>
      <v-toolbar-title class="nsys-toolbar-icon">
        <img src="../assets/icom_logo.svg" alt>
      </v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn icon>
        <v-icon>notifications_none</v-icon>
      </v-btn>
      <v-btn icon>
        <v-icon>person_outline</v-icon>
      </v-btn>
      <v-btn icon>
        <v-icon>help_outline</v-icon>
      </v-btn>
    </v-toolbar>
  </div>
</template>

<script>
export default {
  name: "app-toolbar",
  data: function() {
    return {
      menu: false
    };
  }
};
</script>

<style scoped>
.nsys-toolbar {
  background-color: #ececec;
  box-shadow: 0 1px #e0e0e0 !important;
}
.nsys-toolbar-icon {
  margin-top: 8px;
  margin-left: -120px;
}
.nsys-toolbar-icon > img {
  height: 48px;
}
.nsys-toolbar-title {
  margin-left: -24px !important;
  padding-left: 24px;
  font-size: 1.25rem;
  padding-right: 8px;
  border-right: 1px solid #e0e0e0;
  cursor: pointer;
  margin-top: 0;
  padding-top: 20px;
  height: 64px;
}
.nsys-toolbar-title:hover {
  background-color: #cfe2e8;
}
.nsys-vcard-title {
  padding-top: 20px !important;
  padding-left: 24px !important;
  height: 64px;
}
.nsys-menu-title {
  font-size: 1.25rem;
  font-weight: 600;
  letter-spacing: 0.02em;
}
/* Vuetify overrides */
.theme--light.v-btn {
  color: #00a5db;
}
.v-menu__content {
  top: 0px !important;
  left: 0px !important;
  box-shadow: none !important;
  -webkit-box-shadow: none !important;
  border-top: 0;
  border-left: 0;
  border-bottom: 1px solid #e0e0e0;
  border-right: 1px solid #e0e0e0;
}
</style>
