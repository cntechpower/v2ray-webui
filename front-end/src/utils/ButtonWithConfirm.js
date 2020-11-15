import React from "react";
import { Modal, Button } from "antd";

const { confirm } = Modal;

class ButtonWithConfirm extends React.Component {
  showConfirm(props) {
    confirm({
      title: props.confirmTitle,
      content: props.confirmContent,
      onOk() {
        props.fnOnOk();
      },
      onCancel() {},
    });
  }

  render() {
    return (
      <Button
        disabled={this.props.btnDisabled}
        loading={this.props.loading}
        onClick={() => this.showConfirm(this.props)}
      >
        {this.props.btnName}
      </Button>
    );
  }
}

export default ButtonWithConfirm;
