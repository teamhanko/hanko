import React, { useCallback, useEffect, useState } from "react";
import { createHankoClient, register } from "@teamhanko/hanko-elements";
import styles from "./Todo.module.css";
import { useNavigate } from "react-router-dom";

const api = process.env.REACT_APP_HANKO_API!

function HankoProfile() {
  const navigate = useNavigate();
  const hankoClient = createHankoClient(api);
  const [error, setError] = useState<Error | null>(null);

  const logout = () => {
    hankoClient.user
      .logout()
      .catch(setError);
  };

  const redirectToTodo = () => {
    navigate("/todo");
  }

  const redirectToLogin = useCallback(() => {
    navigate("/");
  }, [navigate]);

  useEffect(() => {
    register(api).catch(setError);
  }, []);

  useEffect(() => hankoClient.onSessionRemoved(() => {
    redirectToLogin();
  }), [hankoClient, redirectToLogin])

  return (
    <>
      <nav className={styles.nav}>
        <button onClick={logout} className={styles.button}>
          Logout
        </button>
        <button disabled className={styles.button}>
          Profile
        </button>
        <button onClick={redirectToTodo} className={styles.button}>
          Todos
        </button>
      </nav>
      <div className={styles.content}>
        <h1 className={styles.headline}>Profile</h1>
        <div className={styles.error}>{error?.message}</div>
        <hanko-profile />
      </div>
    </>
  );
}

export default HankoProfile;
