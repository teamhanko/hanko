import React, { useMemo, useState } from "react";
import { NextPage } from "next";
import { useRouter } from "next/router";
import { TodoClient } from "../util/TodoClient";
import styles from "../styles/Todo.module.css";
import dynamic from "next/dynamic";

const todoApi = process.env.NEXT_PUBLIC_TODO_API!;

const HankoProfile = dynamic(() => import("../components/HankoProfile"), {
  ssr: false,
});

const Todo: NextPage = () => {
  const router = useRouter();
  const client = useMemo(() => new TodoClient(todoApi), []);
  const [error, setError] = useState<Error | null>(null);

  const logout = () => {
    client
      .logout()
      .then(() => {
        router.push("/").catch((e) => setError(e));
        return;
      })
      .catch((e) => {
        setError(e);
      });
  };

  const todos = () => {
    router.push("/todo").catch((e) => setError(e));
  };

  return (
    <>
      <nav className={styles.nav}>
        <button onClick={logout} className={styles.button}>
          Logout
        </button>
        <button disabled className={styles.button}>
          Profile
        </button>
        <button onClick={todos} className={styles.button}>
          Todos
        </button>
      </nav>
      <div className={styles.content}>
        <h1 className={styles.headline}>Profile</h1>
        <div className={styles.error}>{error?.message}</div>
        <HankoProfile setError={setError} />
      </div>
    </>
  );
};

export default Todo;
