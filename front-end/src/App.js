import React from "react";
import "antd/dist/antd.css";
import "./App.css";
import { Layout, Menu } from "antd";
import { Link } from "react-router-dom";
import {
  DesktopOutlined,
  PieChartOutlined,
  GlobalOutlined,
} from "@ant-design/icons";

const { Header, Content, Footer, Sider } = Layout;
const { SubMenu } = Menu;

class SiderBar extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      collapsed: false,
    };
  }

  onCollapse = (collapsed) => {
    console.log(collapsed);
    this.setState({ collapsed });
  };

  render() {
    const { collapsed } = this.state;
    const selectedKey = this.props.selectKey || "";
    const openKey = this.props.openKey || "";
    return (
      <Layout style={{ minHeight: "100vh" }}>
        <Sider collapsible collapsed={collapsed} onCollapse={this.onCollapse}>
          <div className="logo" onClick={this.goHome}>
            <h3 class="h3">V2ray管理平台</h3>
          </div>
          <Menu
            theme="dark"
            defaultSelectedKeys={new Array(selectedKey)}
            defaultOpenKeys={new Array(openKey)}
            mode="inline"
          >
            <Menu.Item key="1" icon={<PieChartOutlined />}>
              系统状态
            </Menu.Item>
            <SubMenu key="pac" icon={<DesktopOutlined />} title="PAC管理">
              <Menu.Item key="pac_config">
                <Link to="/pac/config">基本配置</Link>
              </Menu.Item>
              <Menu.Item key="pac_websites">
                <Link to="/pac/websites">网址管理</Link>
              </Menu.Item>
            </SubMenu>
            <SubMenu key="v2ray" icon={<GlobalOutlined />} title="V2ray管理">
              <Menu.Item key="v2ray_subs">
                <Link to="/v2ray/subscriptions">订阅管理</Link>
              </Menu.Item>
              <Menu.Item key="v2ray_servers">
                <Link to="/v2ray/servers">节点列表</Link>
              </Menu.Item>
              <Menu.Item key="v2ray_config">
                <Link to="/v2ray/config">配置模板</Link>
              </Menu.Item>
            </SubMenu>
          </Menu>
        </Sider>
        <Layout className="site-layout">
          <Header className="site-layout-background" style={{ padding: 0 }} />
          <Content style={{ margin: "0 16px" }}>
            <div
              className="site-layout-background"
              style={{ padding: 24, minHeight: 360 }}
            >
              {this.props.children}
            </div>
          </Content>
          <Footer style={{ textAlign: "center" }}>
            Ant Design ©2018 Created by Ant UED
          </Footer>
        </Layout>
      </Layout>
    );
  }
}

export default SiderBar;
