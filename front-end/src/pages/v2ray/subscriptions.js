import React from "react";
import "antd/dist/antd.css";
// import "./index.css";
import ButtonWithConfirm from "../../utils/ButtonWithConfirm";
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
import { SyncOutlined, PlusOutlined } from "@ant-design/icons";
import axios from "axios";
import Draggable from "react-draggable";
import api from "../../api/api.js";

class V2raySubscriptions extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      error: null,
      isLoaded: false,
      modalVisible: false,
      data: null,
    };
  }
  refreshV2raySubscriptionsList = () => {
    var self = this;
    axios
      .get(api.refreshV2raySubscriptionsListApi)
      .then(function (response) {
        self.setState({
          isLoaded: true,
          data: response.data,
        });
      })
      .catch(function (error) {
        self.setState({
          isLoaded: true,
          error,
        });
      });
  };
  componentDidMount() {
    this.refreshV2raySubscriptionsList();
  }

  showAddModal = () => {
    this.setState({
      modalVisible: true,
    });
  };

  hideAddModal = () => {
    this.setState({
      modalVisible: false,
    });
  };

  openNotificationWithIcon = (type, title, message) => {
    notification[type]({
      message: title,
      description: message,
    });
  };

  addV2raySubscriptions = (name, addr) => {
    var self = this;
    var data = new FormData();
    data.append("subscription_name", name);
    data.append("subscription_addr", addr);
    axios
      .post(api.addV2raySubscriptionsApi, data)
      .then(function (response) {
        self.openNotificationWithIcon(
          "success",
          "添加订阅成功",
          "成功添加订阅: " + name
        );
      })
      .catch(function (error) {
        self.openNotificationWithIcon(
          "error",
          "添加订阅失败",
          "添加订阅 " + name + " 失败. " + error.response.data.Message
        );
      })
      .then(function () {
        self.refreshV2raySubscriptionsList();
        self.hideAddModal();
      });
  };

  delV2raySubscriptions = (name, id) => {
    var self = this;
    var data = new FormData();
    data.append("subscription_id", id);
    axios
      .post(api.delPacWebsiteApi, data)
      .then(function (response) {
        self.openNotificationWithIcon(
          "success",
          "删除订阅成功",
          "成功删除订阅: " + name
        );
      })
      .catch(function (error) {
        self.openNotificationWithIcon(
          "error",
          "删除订阅失败",
          "删除订阅失败. " + error.response.data.Message
        );
      })
      .then(function () {
        self.refreshV2raySubscriptionsList();
      });
  };

  refreshV2raySubscriptions = (name, id) => {
    var self = this;
    var data = new FormData();
    data.append("subscription_id", id);
    axios
      .post(api.refreshV2raySubscriptionsApi, data)
      .then(function (response) {
        self.openNotificationWithIcon(
          "success",
          "刷新订阅成功",
          "成功刷新订阅: " + name
        );
      })
      .catch(function (error) {
        self.openNotificationWithIcon(
          "error",
          "刷新订阅失败",
          "刷新订阅失败. " + error.response.data.Message
        );
      })
      .then(function () {
        self.refreshV2raySubscriptionsList();
      });
  };

  render() {
    const columns = [
      {
        title: "订阅ID",
        dataIndex: "id",
        key: "id",
        sorter: (a, b) => a.id - b.id,
      },
      {
        title: "订阅别名",
        dataIndex: "subscription_name",
        key: "subscription_name",
        sorter: (a, b) =>
          a.subscription_name.length - b.subscription_name.length,
      },
      {
        title: "订阅地址",
        dataIndex: "subscription_addr",
        key: "subscription_addr",
      },
      {
        title: "Action",
        key: "action",
        render: (text, record) => (
          <ButtonWithConfirm
            btnName="删除"
            confirmTitle="是否删除此订阅?"
            confirmContent={record.id + " : " + record.subscription_name}
            fnOnOk={() =>
              this.delV2raySubscriptions(record.subscription_name, record.id)
            }
          />
        ),
      },
      {
        title: "Action",
        key: "action",
        render: (text, record) => (
          <ButtonWithConfirm
            btnName="刷新节点"
            confirmTitle="是否刷新此订阅节点?"
            confirmContent={record.id + " : " + record.subscription_name}
            fnOnOk={() =>
              this.refreshV2raySubscriptions(
                record.subscription_name,
                record.id
              )
            }
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
              icon={<SyncOutlined />}
              onClick={this.refreshV2raySubscriptionsList}
            >
              刷新
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={this.showAddModal}
            >
              添加订阅
            </Button>
            <Modal
              title={
                <div
                  style={{
                    width: "100%",
                    cursor: "move",
                  }}
                >
                  添加订阅
                </div>
              }
              visible={this.state.modalVisible}
              footer={[
                <Button
                  form="addV2raySubscriptionForm"
                  key="submit"
                  htmlType="submit"
                >
                  Submit
                </Button>,
              ]}
              onCancel={this.hideAddModal}
              modalRender={(modal) => <Draggable>{modal}</Draggable>}
            >
              <Form
                id="addV2raySubscriptionForm"
                form={form}
                layout="vertical"
                initialValues={{ modifier: "public" }}
                onFinish={(values) => {
                  this.addV2raySubscriptions(values.name, values.addr);
                }}
              >
                <Form.Item
                  name="name"
                  label="订阅别名"
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
                  name="addr"
                  label="订阅地址"
                  rules={[
                    {
                      required: true,
                      type: "url",
                      message: "请输入正确的订阅地址!",
                    },
                  ]}
                >
                  <Input />
                </Form.Item>
              </Form>
            </Modal>
          </Space>
          <Divider />
          <Table columns={columns} dataSource={data} loading={!isLoaded} />;
        </>
      );
    }
  }
}

export default V2raySubscriptions;
