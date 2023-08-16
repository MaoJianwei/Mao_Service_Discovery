const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  productionSourceMap: false,
  transpileDependencies: true,
  devServer: {
    host: '192.168.1.101',
    port: 8080,  //没被占用，可以使用的端口
    open: false,
    proxy: {
      '/api': {
        target: 'http://pi-dpdk.maojianwei.com:29999/', //接口域名
        changeOrigin: true,             //是否跨域
        secure: false,                   //是否https接口
        pathRewrite: {                  //路径重置
          '^/api': ''
        }
      }
    }
  }
})
