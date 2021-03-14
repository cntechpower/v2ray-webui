import React from "react";
import "antd/dist/antd.css";
import api from "../../api/api.js";
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

class PacWebSites extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      error: null,
      isLoaded: false,
      modalVisible: false,
      pacRefreshing: false,
      data: null,
    };
  }
  refreshPacWebsitesList = () => {
    var self = this;
    axios
      .get(api.refreshPacWebsitesListApi)
      .then(function (response) {
        self.setState({
          isLoaded: true,
          data: response.data.data,
        });
      })
      .catch(function (error) {
        self.setState({
          isLoaded: true,
          error,
        });
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
    this.refreshPacWebsitesList();
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

  addPacWebsite = (addr) => {
    var self = this;
    var data = new FormData();
    data.append("web_site", addr);
    axios
      .post(api.addPacWebsiteApi, data)
      .then(function (response) {
        self.openNotificationWithIcon(
          "success",
          "添加网址成功",
          "成功添加网址: " + addr
        );
        self.refreshPacWebsitesList();
      })
      .catch(function (error) {
        self.openNotificationWithIcon(
          "error",
          "添加网址失败",
          "添加网址 " + addr + " 失败. " + error.response.data.message
        );
        self.refreshPacWebsitesList();
      })
      .then(function () {
        self.hideAddModal();
      });
  };

  delPacWebsite = (id, name) => {
    var self = this;
    var data = new FormData();
    data.append("website_id", id);
    axios
      .post(api.delPacWebsiteApi, data)
      .then(function (response) {
        self.openNotificationWithIcon(
          "success",
          "删除网址成功",
          "成功删除网址: " + name
        );
        self.refreshPacWebsitesList();
      })
      .catch(function (error) {
        self.openNotificationWithIcon(
          "error",
          "删除网址失败",
          "删除网址 " + id + " 失败. " + error.response.data.message
        );
        self.refreshPacWebsitesList();
      });
  };

  render() {
    const columns = [
      {
        title: "网址ID",
        dataIndex: "id",
        key: "id",
      },
      {
        title: "网址URL",
        dataIndex: "url",
        key: "url",
      },
      {
        title: "Action",
        key: "action",
        render: (text, record) => (
          <ButtonWithConfirm
            btnName="删除"
            confirmTitle="是否删除此网址?"
            confirmContent={record.url}
            fnOnOk={() => this.delPacWebsite(record.id, record.url)}
          />
        ),
      },
    ];
    const { error, isLoaded, data } = this.state;
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
              onClick={this.refreshPacWebsitesList}
            >
              刷新
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={this.showAddModal}
            >
              添加网址
            </Button>
            <Button
              type="primary"
              loadings={this.state.pacRefreshing}
              icon={<SyncOutlined />}
              onClick={this.refreshPac}
            >
              触发PAC生成
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
                <Button form="addPacWebsiteForm" key="submit" htmlType="submit">
                  Submit
                </Button>,
              ]}
              onCancel={this.hideAddModal}
              modalRender={(modal) => <Draggable>{modal}</Draggable>}
            >
              <Form
                id="addPacWebsiteForm"
                layout="vertical"
                initialValues={{ modifier: "public" }}
                onFinish={(values) => {
                  this.addPacWebsite(values.addr);
                }}
              >
                <Form.Item
                  name="addr"
                  label="地址"
                  rules={[
                    {
                      required: true,
                      message: "请输入URL地址!",
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

export default PacWebSites;
