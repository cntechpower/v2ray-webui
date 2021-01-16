import React from "react";
import "antd/dist/antd.css";
// import "./index.css";
import ButtonWithConfirm from "../../utils/ButtonWithConfirm";
import Draggable from "react-draggable";
import {
  Table,
  Result,
  notification,
  Button,
  Space,
  Modal,
  Form,
  Input,
  Divider,
} from "antd";
import { CloudSyncOutlined, PlusOutlined } from "@ant-design/icons";
import axios from "axios";
import api from "../../api/api.js";

class V2rayNodes extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      error: null,
      isLoaded: false,
      nodePingDoing: false,
      items: [],
    };
  }
  refreshV2rayNodeList = () => {
    var self = this;
    axios
      .get(api.refreshV2rayNodeListApi)
      .then(function (response) {
        self.setState({
          isLoaded: true,
          data: response.data,
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
  componentDidMount() {
    this.refreshV2rayNodeList();
  }

  openNotificationWithIcon = (type, title, message) => {
    notification[type]({
      message: title,
      description: message,
    });
  };

  showAddModal = () => {
    this.setState({
      modalVisible: true,
    });
  };

  hideAddModal = () => {
    this.setState({
      modalVisible: false,
      form: null,
    });
  };

  addV2rayNode(nodeName, nodeHost, nodePort, nodePath, nodeServerId) {
    var self = this;
    var data = new FormData();
    data.append("name", nodeName);
    data.append("host", nodeHost);
    data.append("port", nodePort);
    data.append("path", nodePath);
    data.append("server_id", nodeServerId);
    axios
      .post(api.addV2rayManualNodeApi, data)
      .then(function (response) {
        self.openNotificationWithIcon(
          "success",
          "添加成功",
          "成功添加节点: " + nodeName
        );
      })
      .catch(function (error) {
        self.openNotificationWithIcon(
          "error",
          "添加失败",
          "添加节点 " + nodeName + " 失败. " + error.response.data.Message
        );
      })
      .then(function () {
        self.refreshV2rayNodeList();
        self.hideAddModal();
      });
  }

  switchV2rayNode(nodeName, nodeId) {
    var self = this;
    var data = new FormData();
    data.append("node_id", nodeId);
    axios
      .post(api.switchV2rayNodeApi, data)
      .then(function (response) {
        self.openNotificationWithIcon(
          "success",
          "切换成功",
          "成功切换至节点: " + nodeName
        );
      })
      .catch(function (error) {
        self.openNotificationWithIcon(
          "error",
          "切换失败",
          "切换至节点 " + nodeName + " 失败. " + error.response.data.Message
        );
      });
  }

  pingAllV2rayNode = () => {
    var self = this;
    self.openNotificationWithIcon(
      "info",
      "测速中",
      "测速需要一定时间, 请耐心等待"
    );
    self.setState({ nodePingDoing: true });
    axios
      .post(api.pingAllV2rayNodeApi)
      .then(function (response) {
        self.openNotificationWithIcon("success", "测速成功", "");
      })
      .catch(function (error) {
        self.openNotificationWithIcon(
          "error",
          "测速失败",
          error.response.data.Message
        );
      })
      .then(function () {
        self.refreshV2rayNodeList();
        self.setState({ nodePingDoing: false });
      });
  };

  render() {
    const columns = [
      {
        title: "节点ID",
        dataIndex: "primary_key",
        key: "primary_key",
        sorter: (a, b) => a.primary_key - b.primary_key,
      },
      {
        title: "所属订阅",
        dataIndex: "subscription_name",
        key: "subscription_name",
      },
      {
        title: "节点名",
        dataIndex: "ps",
        key: "ps",
        sorter: (a, b) => a.ps.length - b.ps.length,
      },
      {
        title: "节点延时(ms)",
        dataIndex: "ping_rtt",
        key: "ping_rtt",
        sorter: (a, b) => a.ping_rtt - b.ping_rtt,
      },
      {
        title: "节点地址",
        dataIndex: "host",
        key: "host",
        sorter: (a, b) => a.host.length - b.host.length,
      },
      {
        title: "节点端口",
        dataIndex: "port",
        key: "port",
        sorter: (a, b) => a.port - b.port,
      },
      {
        title: "Action",
        key: "action",
        render: (text, record) => (
          <ButtonWithConfirm
            btnName="使用"
            confirmTitle="是否使用此节点?"
            confirmContent={record.subscription_id + " : " + record.ps}
            fnOnOk={() => this.switchV2rayNode(record.ps, record.primary_key)}
          />
        ),
      },
    ];
    const { error, isLoaded, data, form } = this.state;
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
              icon={<CloudSyncOutlined />}
              onClick={this.pingAllV2rayNode}
              loading={this.state.nodePingDoing}
            >
              节点测速
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={this.showAddModal}
            >
              添加节点
            </Button>
            <Modal
              title={
                <div
                  style={{
                    width: "100%",
                    cursor: "move",
                  }}
                >
                  添加节点
                </div>
              }
              visible={this.state.modalVisible}
              footer={[
                <Button form="addV2rayNodeForm" key="submit" htmlType="submit">
                  Submit
                </Button>,
              ]}
              onCancel={this.hideAddModal}
              modalRender={(modal) => <Draggable>{modal}</Draggable>}
            >
              <Form
                id="addV2rayNodeForm"
                form={form}
                layout="vertical"
                initialValues={{ modifier: "public" }}
                onFinish={(values) => {
                  this.addV2rayNode(
                    values.name,
                    values.host,
                    values.port,
                    values.path,
                    values.server_id
                  );
                }}
              >
                <Form.Item
                  name="name"
                  label="节点名"
                  rules={[
                    {
                      required: true,
                      message: "请输入订阅别名!",
                    },
                  ]}
                >
                  <Input />
                </Form.Item>
                <Form.Item
                  name="host"
                  label="节点地址"
                  rules={[
                    {
                      required: true,
                      message: "请输入正确的节点地址!",
                    },
                  ]}
                >
                  <Input />
                </Form.Item>
                <Form.Item
                  name="port"
                  label="节点端口"
                  rules={[
                    {
                      required: true,
                      message: "请输入正确的节点端口!",
                    },
                  ]}
                >
                  <Input />
                </Form.Item>
                <Form.Item
                  name="path"
                  label="节点Path"
                  rules={[
                    {
                      required: true,
                      message: "请输入正确的节点path!",
                    },
                  ]}
                >
                  <Input />
                </Form.Item>
                <Form.Item
                  name="server_id"
                  label="节点server id"
                  rules={[
                    {
                      required: true,
                      message: "请输入正确的节点server_id!",
                    },
                  ]}
                >
                  <Input />
                </Form.Item>
              </Form>
            </Modal>
          </Space>
          <Divider />
          <Table columns={columns} dataSource={data} loading={!isLoaded} />
        </>
      );
    }
  }
}

export default V2rayNodes;
