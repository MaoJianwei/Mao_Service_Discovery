<template>
  <div id="app">
    <el-container class="class-container">
      <el-header style="font-size: xx-large;">Bigmao Radio Station</el-header>
      <el-container>
        <el-aside style="width: 210px">
          <el-menu default-active="/deviceStatus" style="height: 100%" router>
<!--            @select="handleSelect"-->

            <el-menu-item index="/deviceStatus">
              <el-icon><Menu /></el-icon>
              设备状态
            </el-menu-item>

            <el-menu-item index="/config">
              <el-icon><Menu /></el-icon>
              服务配置
            </el-menu-item>

            <el-menu-item index="/topo">
              <el-icon><Menu /></el-icon>
              数字地图
            </el-menu-item>

            <el-menu-item index="/apiListOldApi">
              <el-icon><Menu /></el-icon>
              API列表、老版WebUI
            </el-menu-item>

<!--            <el-menu-item index="/ONOS">-->
<!--              <el-icon>-->
<!--                <Loading />-->
<!--              </el-icon>-->
<!--              ONOS-->
<!--            </el-menu-item>-->
          </el-menu>
        </el-aside>

        <el-main>
          <router-view name="maoMain"></router-view>
        </el-main>
      </el-container>
    </el-container>

  </div>
</template>

<script>

export default {
  name: 'App',
  components: {
  },
  methods: {
    // handleSelect(index, indexPath) {
    //   if (index === "/config") {
    //     console.log(index, indexPath)
    //   }
    // }
  }
}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  /*margin-top: 60px;*/
}
.class-container {
  /*background:url('../img/bg.jpg') repeat center top #826e79;*/
  height: 100%;
  width: 100%;
  position: fixed;
}
</style>
