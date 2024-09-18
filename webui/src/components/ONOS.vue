<template>
  <!--  <div id="onosOrTips" style="height: 100%; width: 100%;"></div>-->

  <el-text id="onosTips" :type="tipsType" size="large"></el-text>

  <!--<iframe src="https://dev.maojianwei.com:4433/onos/ui" width="100%" height="100%"></iframe>-->
</template>



<script>
export default {
  name: "ONOS",

  data() {
    return {
      tipsType:"danger",
    };
  },

  mounted() {
    this.queryOnosPage();
  },

  methods: {
    queryOnosPage() {
      var vueThis = this;
      this.$http.get("/api/getOnosInfo")
          .then(function (res) {

            var addrPort = res.data["addrPort"];
            if (addrPort != null && addrPort !== "") {
              vueThis.tipsType = "success";
              document.getElementById("onosTips").innerText = "在新窗口中查看拓扑";

              window.open("http://" + addrPort + "/onos/ui", "_blank");
            } else {
              vueThis.tipsType = "warning";
              document.getElementById("onosTips").innerText = "拓扑接口暂未配置";
            }
          })
          .catch(function (err) {
            console.log("errMao: " + err);
            vueThis.tipsType = "danger";
            document.getElementById("onosTips").innerText = "拓扑接口获取异常：" + err;
          });
    },
  }
}
</script>