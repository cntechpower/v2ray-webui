/* eslint-disable import/no-anonymous-default-export */
let prefix = "";
// eslint-disable-next-line no-undef
if (process.env.NODE_ENV === "development") {
  prefix = "http://127.0.0.1:8888";
}

export default {
  //pac pages
  refreshPacWebsitesListApi: prefix + "/api/pac/website/list",
  addPacWebsiteApi: prefix + "/api/pac/website/add",
  delPacWebsiteApi: prefix + "/api/pac/website/del",

  //v2ray pages
  refreshV2rayNodeListApi: prefix + "/api/v2ray/subscription/nodes/list",
  switchV2rayNodeApi: prefix + "/api/v2ray/config/switch_node",
  pingAllV2rayNodeApi: prefix + "/api/v2ray/subscription/nodes/ping",
  refreshV2raySubscriptionsListApi: prefix + "/api/v2ray/subscription/list",
  addV2raySubscriptionsApi: prefix + "/api/v2ray/subscription/add",
  delV2raySubscriptionsApi: prefix + "/api/v2ray/subscription/delete",
  refreshV2raySubscriptionsApi: prefix + "/api/v2ray/subscription/refresh",
};
