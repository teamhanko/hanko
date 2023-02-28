import React, {useCallback, useEffect, useState} from "react";
import { useNavigate } from "react-router-dom";
import { register } from "@teamhanko/hanko-elements";
import styles from "./Todo.module.css";

const api = process.env.REACT_APP_HANKO_API!;

function HankoAuth() {
  const navigate = useNavigate();
  const [error, setError] = useState<Error | null>(null);

  const redirectToTodos = useCallback(() => {
    navigate("/todo", { replace: true });
  }, [navigate]);

  useEffect(() => {
    register({ shadow: true }).catch(setError);
    document.addEventListener("hankoAuthSuccess", redirectToTodos);
    return () =>
      document?.removeEventListener("hankoAuthSuccess", redirectToTodos);
  }, [redirectToTodos, setError]);

  return (
    <div className={styles.content}>
      <div className={styles.error}>{error?.message}</div>
      <hanko-auth api={api} />
    </div>
  );
}

export default HankoAuth;
