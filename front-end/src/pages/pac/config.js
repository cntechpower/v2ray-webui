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
      pacRefreshing: false,
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
  refreshPacConfig = () => {
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
  updatePacConfig = (cron, cmd) => {
    var self = this;
    var data = new FormData();
    data.append("cron", cron);
    data.append("cmd", cmd);
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
          "更新配置失败. " + error.response.data.message
        );
        self.refreshV2rayConfig();
      });
  };

  refreshPac = () => {
    var self = this;
    self.openNotificationWithIcon(
      "info",
      "更新PAC中",
      "更新需要一定时间, 请耐心等待"
    );
    self.setState({ pacRefreshing: true });
    axios
      .post(api.refreshPacApi)
      .then(function (response) {
        self.openNotificationWithIcon("success", "更新成功", "");
      })
      .catch(function (error) {
        self.openNotificationWithIcon(
          "error",
          "更新失败",
          error.response.data.message
        );
      })
      .then(function () {
        self.setState({ pacRefreshing: false });
      });
  };
  componentDidMount() {
    this.refreshPacConfig();
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
              onClick={this.refreshPacConfig}
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
            <Button
              type="primary"
              loadings={this.state.pacRefreshing}
              icon={<SyncOutlined />}
              onClick={this.refreshPac}
            >
              触发PAC生成
            </Button>
          </Space>
          <Divider />
          <Form
            id="updateV2rayConfigForm"
            layout="vertical"
            initialValues={{ modifier: "public" }}
            onFinish={(values) => {
              this.updatePacConfig(values.cron, values.cmd);
            }}
          >
            <Form.Item
              name="cron"
              label="自动更新周期Cron"
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
              name="cmd"
              label="更新PAC命令模板"
              initialValue={this.state.config.cmd}
              rules={[
                {
                  required: true,
                  message: "请输入命令模板!",
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
