import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import HankoAuth from "./HankoAuth";
import Todo from "./Todo";
import "./index.css";
import HankoProfile from "./HankoProfile";

const root = ReactDOM.createRoot(
  document.getElementById("root") as HTMLElement
);

root.render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HankoAuth />} />
        <Route path="/todo" element={<Todo />} />
        <Route path="/profile" element={<HankoProfile />} />
      </Routes>
    </BrowserRouter>
  </React.StrictMode>
);
