import React from "react";
import "antd/dist/antd.css";
import {
  Skeleton,
  Result,
  Button,
  notification,
  Space,
  Divider,
  Tooltip,
  Popover,
} from "antd";
import {
  SyncOutlined,
  BugOutlined,
  SaveOutlined,
  QuestionOutlined,
} from "@ant-design/icons";
import { Controlled as CodeMirror } from "react-codemirror2";
import axios from "axios";
import api from "../../api/api.js";

import "codemirror/lib/codemirror.css";
import "codemirror/theme/material.css";
import "codemirror/theme/neat.css";
import "codemirror/mode/xml/xml.js";
import "codemirror/mode/javascript/javascript.js";

class V2rayConfig extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      error: null,
      isLoaded: false,
      configValidate: false,
      config: null,
    };
  }

  refreshV2rayConfig = () => {
    var self = this;
    axios
      .get(api.refreshV2rayConfigApi)
      .then(function (response) {
        self.setState({
          isLoaded: true,
          config: response.data.data,
        });
      })
      .catch(function (error) {
        console.log(error);
        self.setState({
          isLoaded: true,
          error,
        });
      })
      .then(function () {
        self.setState({ configValidate: false });
      });
  };

  openNotificationWithIcon = (type, title, message) => {
    notification[type]({
      message: title,
      description: message,
    });
  };

  updateV2rayConfig = (config) => {
    var self = this;
    var data = new FormData();
    data.append("config_content", JSON.stringify(self.state.config));
    axios
      .post(api.updateV2rayConfigApi, data)
      .then(function (response) {
        self.openNotificationWithIcon(
          "success",
          "修改配置成功",
          "成功修改配置"
        );
        self.refreshV2rayConfig();
      })
      .catch(function (error) {
        self.openNotificationWithIcon(
          "error",
          "修改配置失败",
          "修改配置失败. " + error.response.data.message
        );
        self.refreshV2rayConfig();
      });
  };

  validateV2rayConfig = (config) => {
    var self = this;
    var data = new FormData();
    data.append("config_content", JSON.stringify(self.state.config));
    axios
      .post(api.validateV2rayConfigApi, data)
      .then(function (response) {
        self.openNotificationWithIcon("success", "校验配置成功", "配置合法");
        self.setState({ configValidate: true });
      })
      .catch(function (error) {
        self.openNotificationWithIcon(
          "error",
          "校验配置失败",
          "配置不合法: " + error.response.data.message
        );
      });
  };
  componentDidMount() {
    this.refreshV2rayConfig();
  }
  render() {
    const { error, isLoaded, config } = this.state;
    if (!isLoaded) {
      return <Skeleton active />;
    }
    const helpContent = (
      <div>
        <p>
          支持如下几种变量, 填写变量到配置模板中后,
          会自动替换为当前使用的节点信息:
        </p>
        <ul>
          <li>1. serverHost : 节点地址</li>
          <li>2. serverName : 节点名</li>
          <li>3. 9495945 : 节点端口</li>
          <li>4. serverId : 节点用户ID</li>
        </ul>
      </div>
    );

    if (error != null) {
      return (
        <Result
          status="warning"
          title="There are some problems with your operation."
          extra={
            <Button type="primary" key="console">
              Go Home
            </Button>
          }
        />
      );
    } else {
      return (
        <>
          <Space>
            <Button
              type="primary"
              icon={<SyncOutlined />}
              onClick={this.refreshV2rayConfig}
            >
              刷新
            </Button>
            <Button
              type="primary"
              icon={<BugOutlined />}
              onClick={this.validateV2rayConfig}
            >
              校验
            </Button>
            <Tooltip placement="top" title="请先校验配置~">
              <Button
                type="primary"
                icon={<SaveOutlined />}
                onClick={this.updateV2rayConfig}
                disabled={!this.state.configValidate}
              >
                保存
              </Button>
            </Tooltip>
            <Popover content={helpContent} title="配置说明">
              <Button shape="circle" icon={<QuestionOutlined />}></Button>
            </Popover>
          </Space>
          <Divider />
          <CodeMirror
            value={config}
            options={{
              mode: { name: "javascript", json: true },
              theme: "material",
              lineNumbers: true,
            }}
            editorDidMount={(editor) => {
              editor.setSize("100%", "800px");
            }}
            onBeforeChange={(editor, data, value) => {
              this.setState({ config: value, configValidate: false });
            }}
          />
        </>
      );
    }
  }
}

export default V2rayConfig;
