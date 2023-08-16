<template>
  <div style="margin: 20px" />

  <el-form :model="form" label-width="120px" label-position="top" style="max-width: 600px">
    <el-form-item label="New Services">
      <el-input v-model="form.ipv4v6" placeholder="one line, one receiver" type="textarea" rows="3"/>
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="onSubmit">Add</el-button>
    </el-form-item>
  </el-form>

  <el-table :data="maoIcmpTableData" ref="maoTable" :cell-class-name="tableCellClassName"
            empty-text="暂无数据" max-height="610px">
    <el-table-column label="Control">
      <template #default="scope">
        <el-button size="small" type="danger" @click="handleDelete(scope.$index, scope.row)">Delete</el-button>
      </template>
    </el-table-column>
    <el-table-column label="Device IP" prop="deviceIp" />
    <el-table-column label="Alive" prop="alive" />
    <el-table-column label="Detect Count" prop="Detect_Count" />
    <el-table-column label="Report Count" prop="Report_Count" />
    <el-table-column label="RTT Duration" prop="RTT_Duration" />
    <el-table-column label="Last Seen" prop="Last_Seen" />
    <el-table-column label="Timestamp" prop="RttOutbound_or_Remote_Timestamp" />
  </el-table>

</template>

<script>
export default {
  name: "ConfigIcmp",
  data() {
    return {
      maoIcmpTableData: [],
      refreshTimer: "",
      form: {
        ipv4v6: "",
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

    handleDelete(index, row) {
      var vueThis = this;
      this.$http.post("/api/delServiceIp", {ipv4v6: row.deviceIp},
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
      this.$http.get("/api/showServiceIP")
          .then(function (res) {
            vueThis.maoIcmpTableData = [];

            var data = res.data;
            for (var i = 0; i < data.length; i++) {
              vueThis.maoIcmpTableData.push(
                  {
                    deviceIp: data[i]["Address"] != null ? data[i]["Address"] : data[i]["Hostname"],
                    alive: data[i]["Alive"],
                    Detect_Count: data[i]["DetectCount"] != null ? data[i]["DetectCount"] : "/",
                    Report_Count: data[i]["ReportCount"] != null ? data[i]["ReportCount"] : data[i]["ReportTimes"],
                    RTT_Duration: data[i]["RttDuration"] != null ? (data[i]["RttDuration"] / 1000 / 1000).toFixed(3) + "ms" : "/",
                    Last_Seen: data[i]["LastSeen"] != null ? data[i]["LastSeen"] : data[i]['LocalLastSeen'],
                    RttOutbound_or_Remote_Timestamp: data[i]["RttOutboundTimestamp"] != null ? data[i]["RttOutboundTimestamp"] : data[i]["ServerDateTime"],
                  }
              );
            }
          })
          .catch(function (err) {
            console.log("errMao: " + err);
          });
    },

    onSubmit() {
      var vueThis = this;
      this.$http.post("/api/addServiceIp", this.form,
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
  }
}
</script>