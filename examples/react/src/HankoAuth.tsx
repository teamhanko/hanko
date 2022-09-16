import React, { useCallback, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { register } from "@teamhanko/hanko-elements/hanko-auth";
import styles from "./Todo.module.css";

const api = process.env.REACT_APP_HANKO_API!;
const lang = process.env.REACT_APP_HANKO_LANG;

function HankoAuth() {
  const navigate = useNavigate();

  const redirectToTodos = useCallback(() => {
    navigate("/todo", { replace: true });
  }, [navigate]);

  useEffect(() => {
    register({ shadow: true });
    document.addEventListener("hankoAuthSuccess", redirectToTodos);
    return () =>
      document?.removeEventListener("hankoAuthSuccess", redirectToTodos);
  }, [redirectToTodos]);

  return (
    <div className={styles.content}>
      <hanko-auth api={api} lang={lang} />
    </div>
  );
}

export default HankoAuth;
