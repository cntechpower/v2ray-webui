import React from "react";
import "antd/dist/antd.css";
import {
  Skeleton,
  Result,
  Button,
  notification,
  Space,
  Statistic,
  Row,
  Col,
} from "antd";
import axios from "axios";
import api from "../api/api.js";

class Home extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      error: null,
      isLoaded: false,
      status: null,
    };
  }

  refreshStatus = () => {
    var self = this;
    axios
      .get(api.refreshStatusApi)
      .then(function (response) {
        self.setState({
          isLoaded: true,
          status: response.data,
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

  componentDidMount() {
    this.refreshStatus();
  }
  render() {
    const { error, isLoaded, status } = this.state;
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
          <Space></Space>
          <Row gutter={16}>
            <Col span={12}>
              <Statistic
                title="当前节点"
                value={status.data}
                formatter={(value) =>
                  (value.current_node && value.current_node.ps) || ""
                }
              />
            </Col>
            <Col span={12}>
              <Statistic
                title="节点延迟"
                value={status.data}
                formatter={(value) =>
                  (value.current_node && value.current_node.ping_rtt) || ""
                }
              />
            </Col>
            <Col span={12}>
              <Statistic
                title="Core启动时间"
                value={status.data.v2ray_core.start_time}
                precision={2}
              />
            </Col>
            <Col span={12}>
              <Statistic
                title="状态刷新时间"
                value={status.data.refresh_time}
              />
            </Col>
          </Row>
          {/* <Divider /> */}
        </>
      );
    }
  }
}

export default Home;
