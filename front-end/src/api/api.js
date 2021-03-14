/* eslint-disable import/no-anonymous-default-export */
let prefix = "";
// eslint-disable-next-line no-undef
if (process.env.NODE_ENV === "development") {
  prefix = "http://127.0.0.1:8888";
}

export default {
  //pac pages
  refreshPacApi: prefix + "/api/pac/update",
  refreshPacWebsitesListApi: prefix + "/api/pac/website/list",
  addPacWebsiteApi: prefix + "/api/pac/website/add",
  delPacWebsiteApi: prefix + "/api/pac/website/del",
  getPacConfigApi: prefix + "/api/pac/config/get",
  updatePacConfigApi: prefix + "/api/pac/config/update",

  //v2ray subscription
  refreshV2raySubscriptionsListApi: prefix + "/api/v2ray/subscription/list",
  addV2raySubscriptionsApi: prefix + "/api/v2ray/subscription/add",
  delV2raySubscriptionsApi: prefix + "/api/v2ray/subscription/delete",
  refreshV2raySubscriptionsApi: prefix + "/api/v2ray/subscription/refresh",

  //v2ray nodes
  pingAllV2rayNodeApi: prefix + "/api/v2ray/nodes/ping",
  addV2rayManualNodeApi: prefix + "/api/v2ray/nodes/add",
  refreshV2rayNodeListApi: prefix + "/api/v2ray/nodes/list",

  //v2ray config
  refreshV2rayConfigApi: prefix + "/api/v2ray/config/get",
  updateV2rayConfigApi: prefix + "/api/v2ray/config/update",
  validateV2rayConfigApi: prefix + "/api/v2ray/config/validate",
  switchV2rayNodeApi: prefix + "/api/v2ray/config/switch_node",

  //v2ray status
  refreshStatusApi: prefix + "/api/status/v2ray",
};
