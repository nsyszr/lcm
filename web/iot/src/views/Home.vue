<template>
  <div class="home">
    <v-container fluid class="pa-0">
      <v-layout row wrap fill-height>
        <v-flex xs12 class="nsys-breadcrumb">
          <h3 class="nsys-breadcrumb-title">
            Smart Devices
            <span>{{desserts.length}}</span>
          </h3>
        </v-flex>
        <v-flex xs9 grow style="border-right: 1px solid #e0e0e0;">
          <v-data-table :headers="headers" :items="desserts" class="elevation-0" hide-actions>
            <template v-slot:items="props">
              <td>{{ props.item.name }}</td>
              <td class="text-xs-right">{{ props.item.calories }}</td>
              <td class="text-xs-right">{{ props.item.fat }}</td>
              <td class="text-xs-right">{{ props.item.carbs }}</td>
              <td class="text-xs-right">{{ props.item.protein }}</td>
              <td class="text-xs-right">{{ props.item.iron }}</td>
            </template>
          </v-data-table>
        </v-flex>
        <v-flex xs3 grow>blaaa</v-flex>
      </v-layout>
    </v-container>
  </div>
</template>

<script>
export default {
  name: "home",
  data: function() {
    return {
      headers: [
        {
          text: "Dessert (100g serving)",
          align: "left",
          sortable: false,
          value: "name"
        },
        { text: "Calories", value: "calories" },
        { text: "Fat (g)", value: "fat" },
        { text: "Carbs (g)", value: "carbs" },
        { text: "Protein (g)", value: "protein" },
        { text: "Iron (%)", value: "iron" }
      ],
      desserts: [
        {
          name: "Frozen Yogurt",
          calories: 159,
          fat: 6.0,
          carbs: 24,
          protein: 4.0,
          iron: "1%"
        },
        {
          name: "Ice cream sandwich",
          calories: 237,
          fat: 9.0,
          carbs: 37,
          protein: 4.3,
          iron: "1%"
        },
        {
          name: "Eclair",
          calories: 262,
          fat: 16.0,
          carbs: 23,
          protein: 6.0,
          iron: "7%"
        },
        {
          name: "Cupcake",
          calories: 305,
          fat: 3.7,
          carbs: 67,
          protein: 4.3,
          iron: "8%"
        },
        {
          name: "Gingerbread",
          calories: 356,
          fat: 16.0,
          carbs: 49,
          protein: 3.9,
          iron: "16%"
        },
        {
          name: "Jelly bean",
          calories: 375,
          fat: 0.0,
          carbs: 94,
          protein: 0.0,
          iron: "0%"
        },
        {
          name: "Lollipop",
          calories: 392,
          fat: 0.2,
          carbs: 98,
          protein: 0,
          iron: "2%"
        },
        {
          name: "Honeycomb",
          calories: 408,
          fat: 3.2,
          carbs: 87,
          protein: 6.5,
          iron: "45%"
        },
        {
          name: "Donut",
          calories: 452,
          fat: 25.0,
          carbs: 51,
          protein: 4.9,
          iron: "22%"
        },
        {
          name: "KitKat",
          calories: 518,
          fat: 26.0,
          carbs: 65,
          protein: 7,
          iron: "6%"
        }
      ]
    };
  }
};
</script>

<style scoped>
.nsys-breadcrumb {
  height: 48px;
  line-height: 48px;
  padding-left: 24px !important;
  border-bottom: 1px solid #e0e0e0;
  vertical-align: middle !important;
}
.nsys-breadcrumb-title {
  font-weight: 600;
  font-size: 1rem;
}
.nsys-breadcrumb-title > span {
  color: #999;
}
.theme--light.v-table tbody {
  border-bottom: 0px !important;
}
</style>
