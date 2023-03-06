import React, { useEffect, useMemo, useState } from "react";
import { register } from "@teamhanko/hanko-elements";
import styles from "./Todo.module.css";
import { useNavigate } from "react-router-dom";
import { TodoClient } from "./TodoClient";

const hankoApi = process.env.REACT_APP_HANKO_API!
const todoApi = process.env.REACT_APP_TODO_API!

function HankoProfile() {
  const navigate = useNavigate();
  const client = useMemo(() => new TodoClient(todoApi), []);
  const [error, setError] = useState<Error | null>(null);

  const logout = () => {
    client
      .logout()
      .then(() => {
        navigate("/");
        return;
      })
      .catch(setError);
  };

  const todo = () => {
    navigate("/todo");
  }

  useEffect(() => {
    register({ shadow: true }).catch(setError);
  }, []);

  return (
    <>
      <nav className={styles.nav}>
        <button onClick={logout} className={styles.button}>
          Logout
        </button>
        <button disabled className={styles.button}>
          Profile
        </button>
        <button onClick={todo} className={styles.button}>
          Todos
        </button>
      </nav>
      <div className={styles.content}>
        <h1 className={styles.headline}>Profile</h1>
        <div className={styles.error}>{error?.message}</div>
        <hanko-profile api={hankoApi} />
      </div>
    </>
  );
}

export default HankoProfile;
