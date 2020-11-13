import React from "react";
import "antd/dist/antd.css";
import { Skeleton, Result, Button, notification, Space, Divider } from "antd";
import { SyncOutlined, CheckCircleOutlined } from "@ant-design/icons";
import JSONInput from "react-json-editor-ajrm";
import locale from "react-json-editor-ajrm/locale/en";
import axios from "axios";
import api from "../../api/api.js";

class V2rayConfig extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      error: null,
      isLoaded: false,
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
          config: response.data,
        });
      })
      .catch(function (error) {
        console.log(error);
        self.setState({
          isLoaded: true,
          error,
        });
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
    data.append("config_content", this.state.config);
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
          "修改配置失败. " + error.response.data.Message
        );
        self.refreshV2rayConfig();
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
              icon={<CheckCircleOutlined />}
              onClick={this.updateV2rayConfig}
            >
              提交
            </Button>
          </Space>
          <Divider />
          <JSONInput
            id="a_unique_id"
            placeholder={config}
            locale={locale}
            theme="light_mitsuketa_tribute"
            colors={{
              string: "#DAA520", // overrides theme colors with whatever color value you want
            }}
            onChange={(values) => {
              this.setState({ config: values });
            }}
            height="750px"
            width="700px"
          />
        </>
      );
    }
  }
}

export default V2rayConfig;
