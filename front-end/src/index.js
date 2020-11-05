import React from "react";
import ReactDOM from "react-dom";
import { BrowserRouter as Router, Switch, Route } from "react-router-dom";
import "./index.css";
import App from "./App";
import V2rayNodes from "./pages/v2ray/nodes";
import V2raySubscriptions from "./pages/v2ray/subscriptions";
import PacWebsites from "./pages/pac/websites";
import reportWebVitals from "./reportWebVitals";

ReactDOM.render(
  <React.StrictMode>
    <Router>
      {/* A <Switch> looks through its children <Route>s and
            renders the first one that matches the current URL. */}
      <Switch>
        <Route path="/pac/config">
          <App openKey="pac" selectKey="pac">
            <PacWebsites />
          </App>
        </Route>
        <Route path="/v2ray/servers">
          <App openKey="v2ray" selectKey="v2ray_servers">
            <V2rayNodes />
          </App>
        </Route>
        <Route path="/v2ray/subscriptions/">
          <App openKey="v2ray" selectKey="v2ray_subs">
            <V2raySubscriptions />
          </App>
        </Route>
        <Route path="/">
          <App>home</App>
        </Route>
      </Switch>
    </Router>
  </React.StrictMode>,
  document.getElementById("root")
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
