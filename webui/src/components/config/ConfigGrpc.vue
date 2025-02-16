<template>
  <div style="margin: 20px" />

<!--  <el-form :model="form" label-width="120px" label-position="top" style="max-width: 600px">-->
<!--    <el-form-item label="New Services">-->
<!--      <el-input v-model="form.ipv4v6" placeholder="one line, one receiver" type="textarea" rows="3"/>-->
<!--    </el-form-item>-->
<!--    <el-form-item>-->
<!--      <el-button type="primary" @click="onSubmit">Add</el-button>-->
<!--    </el-form-item>-->
<!--  </el-form>-->

  <el-table :data="maoGrpcTableData" ref="maoTable" :cell-class-name="tableCellClassName"
            empty-text="暂无数据" max-height="610px">
    <el-table-column label="Control">
      <template #default="scope">
        <el-button size="small" type="danger" @click="handleDelete(scope.$index, scope.row)">Delete</el-button>
      </template>
    </el-table-column>


    <el-table-column label="Service Name" prop="serviceName" />
    <el-table-column label="Report IP" prop="deviceIps" />
    <el-table-column label="Alive" prop="alive" />
    <el-table-column label="Report Count" prop="reportCount" />
    <el-table-column label="RTT Duration" prop="rttDuration" />
    <el-table-column label="Last Seen" prop="lastSeen" />
    <el-table-column label="Timestamp" prop="remoteTimestamp" />
  </el-table>

</template>

<script>
export default {
  name: "ConfigGrpc",
  data() {
    return {
      maoGrpcTableData: [],
      refreshTimer: "",
      form: {
        serviceNames: "",
      }
    }
  },

  mounted() {
    this.onLoad()
    this.refreshTimer = setInterval(this.onLoad, 1000);
  },
  beforeUnmount() {
    clearInterval(this.refreshTimer);
  },

  methods: {
    tableCellClassName({row, column, rowIndex, columnIndex}) {
      //利用单元格的 className 的回调方法，给行列索引赋值
      row.index = rowIndex;
      column.index = columnIndex;
    },

    // TODO: TBD
    handleDelete(index, row) {
      var vueThis = this;
      this.$http.post("/api/delGrpcService", {serviceNames: row.deviceIp},
          {
            headers: {
              'Content-Type': 'application/x-www-form-urlencoded;'
            }
          })
          .then(function () { // res
            // setTimeout(vueThis.onLoad, 500)
            vueThis.onLoad()
          })
          .catch(function (err) {
            console.log("errMao: " + err);
          });
    },

    onLoad() {
      var vueThis = this;
      this.$http.get("/api/showAllGrpcService")
          .then(function (res) {
            vueThis.maoGrpcTableData = [];

            var data = res.data;
            for (var i = 0; i < data.length; i++) {
              vueThis.maoGrpcTableData.push(
                  {
                    serviceName: data[i]["Hostname"],
                    deviceIps: data[i]["Ips"].join("\n"),
                    alive: data[i]["Alive"],
                    reportCount: data[i]["ReportTimes"],
                    rttDuration: (data[i]["RttDuration"] / 1000 / 1000).toFixed(3) + "ms",
                    lastSeen: data[i]["LocalLastSeen"],
                    remoteTimestamp: data[i]["ServerDateTime"],
                  }
              );
            }
          })
          .catch(function (err) {
            console.log("errMao: " + err);
          });
    },

    // onSubmit() {
    //   var vueThis = this;
    //   this.$http.post("/api/addServiceIp", this.form,
    //       {
    //         headers: {
    //           'Content-Type': 'application/x-www-form-urlencoded;'
    //         }
    //       })
    //       .then(function () { // res
    //         // setTimeout(vueThis.onLoad, 500)
    //         vueThis.onLoad()
    //       })
    //       .catch(function (err) {
    //         console.log("errMao: " + err);
    //       });
    // },
  }
}
</script>