import React, { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { createHankoClient, register } from "@teamhanko/hanko-elements";
import styles from "./Todo.module.css";

const api = process.env.REACT_APP_HANKO_API!;

function HankoAuth() {
  const navigate = useNavigate();
  const [error, setError] = useState<Error | null>(null);
  const hankoClient = createHankoClient(api);

  const redirectToTodos = useCallback(() => {
    navigate("/todo", { replace: true });
  }, [navigate]);

  useEffect(() => {
    register(api).catch(setError);
  }, []);

  useEffect(() => hankoClient.onAuthFlowCompleted(() => {
    redirectToTodos();
  }), [hankoClient, redirectToTodos])

  return (
    <div className={styles.content}>
      <div className={styles.error}>{error?.message}</div>
      <hanko-auth />
    </div>
  );
}

export default HankoAuth;
