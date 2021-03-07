import React from "react";
import "antd/dist/antd.css";
import api from "../../api/api.js";
import {
  Form,
  Input,
  Skeleton,
  Button,
  Space,
  Divider,
  Result,
  notification,
} from "antd";
import { SyncOutlined, SaveOutlined, EditOutlined } from "@ant-design/icons";
import axios from "axios";
class PacConfig extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      error: null,
      isLoaded: false,
      formDisable: true,
      config: null,
    };
  }
  openNotificationWithIcon = (type, title, message) => {
    notification[type]({
      message: title,
      description: message,
    });
  };
  refreshV2rayConfig = () => {
    var self = this;
    self.setState({
      isLoaded: false,
      formDisable: true,
    });
    axios
      .get(api.getPacConfigApi)
      .then(function (response) {
        self.setState({
          isLoaded: true,
          config: response.data.data,
        });
      })
      .catch(function (error) {
        self.setState({
          isLoaded: true,
          error,
        });
      })
      .then(function () {
        self.setState({
          isLoaded: true,
        });
      });
  };

  updatePacConfig = (config) => {
    console.log(config);
  };

  updatePacConfig = (cron, proxy_addr) => {
    var self = this;
    var data = new FormData();
    data.append("cron", cron);
    data.append("proxy_addr", proxy_addr);
    axios
      .post(api.updatePacConfigApi, data)
      .then(function (response) {
        self.openNotificationWithIcon("success", "更新配置成功");
        self.refreshV2rayConfig();
      })
      .catch(function (error) {
        self.openNotificationWithIcon(
          "error",
          "更新配置失败",
          "更新配置失败. " + error.response.data.Message
        );
        self.refreshV2rayConfig();
      });
  };
  componentDidMount() {
    this.refreshV2rayConfig();
  }
  render() {
    if (!this.state.isLoaded) {
      return <Skeleton active />;
    }
    if (this.state.error != null) {
      return (
        <Result
          status="warning"
          title="There are some problems with your operation."
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
              icon={<EditOutlined />}
              onClick={() => {
                this.setState({ formDisable: false });
              }}
            >
              编辑
            </Button>
            <Button
              type="primary"
              disabled={this.state.formDisable}
              icon={<SaveOutlined />}
              form="updateV2rayConfigForm"
              key="submit"
              htmlType="submit"
            >
              保存
            </Button>
          </Space>
          <Divider />
          <Form
            id="updateV2rayConfigForm"
            layout="vertical"
            initialValues={{ modifier: "public" }}
            onFinish={(values) => {
              this.updatePacConfig(values.cron, values.proxy_addr);
            }}
          >
            <Form.Item
              name="cron"
              label="自动更新Cron -- (会在cron指定的时间自动使用最新的GFW更新PAC)"
              initialValue={this.state.config.cron}
              rules={[
                {
                  required: true,
                  message: "请输入自动更新Cron!",
                },
              ]}
            >
              <Input disabled={this.state.formDisable} />
            </Form.Item>
            <Form.Item
              name="proxy_addr"
              label="PAC代理地址 -- (请与V2ray配置模板中的端口保持一致)"
              initialValue={this.state.config.proxy_addr}
              rules={[
                {
                  required: true,
                  message: "请输入代理地址!",
                },
              ]}
            >
              <Input disabled={this.state.formDisable} />
            </Form.Item>
          </Form>
        </>
      );
    }
  }
}

export default PacConfig;
