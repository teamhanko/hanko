import React, {useEffect, useMemo, useState} from "react";
import { register } from "@teamhanko/hanko-elements";
import styles from "./Todo.module.css";
import {useNavigate} from "react-router-dom";
import {TodoClient} from "./TodoClient";

const api = process.env.REACT_APP_HANKO_API!;

function HankoProfile() {
  const navigate = useNavigate();
  const client = useMemo(() => new TodoClient(api), []);
  const [error, setError] = useState<Error | null>(null);

  const logout = () => {
    client
      .logout()
      .then(() => {
        navigate("/");
        return;
      })
      .catch((e) => {
        setError(e);
      });
  };

  const todo = () => {
    navigate("/todo");
  }

  useEffect(() => {
    register({ shadow: true }).catch(console.error);
  }, []);

  return (
    <>
      <nav className={styles.nav}>
        <button onClick={logout} className={styles.button}>
          Logout
        </button>
        <button onClick={todo} className={styles.button}>
          Todo
        </button>
      </nav>
    <div className={styles.content}>
      <h1 className={styles.headline}>Profile</h1>
      <div className={styles.error}>{error?.message}</div>
      <hanko-profile api={api} />
    </div>
    </>
  );
}

export default HankoProfile;
