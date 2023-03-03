import React, { useCallback, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { register } from "@teamhanko/hanko-elements";
import styles from "./Todo.module.css";

const api = process.env.REACT_APP_HANKO_API!;

function HankoAuth() {
  const navigate = useNavigate();

  const redirectToTodos = useCallback(() => {
    navigate("/todo", { replace: true });
  }, [navigate]);

  useEffect(() => {
    register({ shadow: true }).catch((e) => console.error(e));
    document.addEventListener("hankoAuthSuccess", redirectToTodos);
    return () =>
      document?.removeEventListener("hankoAuthSuccess", redirectToTodos);
  }, [redirectToTodos]);

  return (
    <div className={styles.content}>
      <h1>Hello from {testString}</h1>
      <hanko-auth api={api} />
    </div>
  );
}

export default HankoAuth;
