// import Vue from "vue"
import { createRouter, createWebHashHistory } from 'vue-router'
import ConfigPages from "@/components/ConfigPages.vue"
import ConfigEmail from "@/components/config/ConfigEmail.vue";
import ConfigIcmp from "@/components/config/ConfigIcmp.vue";
import ConfigTopo from "@/components/config/ConfigTopo.vue";

import DeviceInfo from "@/components/DeviceInfo.vue";
import ONOS from "@/components/ONOS";
import ApiListOldApi from "@/components/ApiListOldApi.vue"

// Vue.use(VueRouter)

export default createRouter({
        history: createWebHashHistory(),
        routes: [
            {
                path: '/',
                components: {
                    maoMain: DeviceInfo,
                }
            },
            {
                path: '/config',
                components: {
                    maoMain: ConfigPages,
                },
                children: [
                    {
                        path: '',
                        components: {
                            maoConfig: ConfigEmail
                        }
                    },
                    {
                        path: 'email',
                        components: {
                            maoConfig: ConfigEmail
                        }
                    },
                    {
                        path: 'icmp',
                        components: {
                            maoConfig: ConfigIcmp
                        }
                    },
                    {
                        path: 'topo',
                        components: {
                            maoConfig: ConfigTopo
                        }
                    }
                ]
            },
            {
                path: '/deviceStatus',
                components: {
                    maoMain: DeviceInfo,
                }
            },
            {
                path: '/onos',
                components: {
                    maoMain: ONOS
                }
            },
            {
                path: '/apiListOldApi',
                components: {
                    maoMain: ApiListOldApi
                }
            }
        ]
    }
)