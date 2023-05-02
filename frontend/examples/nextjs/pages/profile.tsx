import React, { useMemo, useState } from "react";
import { NextPage } from "next";
import { useRouter } from "next/router";
import styles from "../styles/Todo.module.css";
import dynamic from "next/dynamic";
import { createHankoClient } from "@teamhanko/hanko-elements";

const hankoAPI = process.env.NEXT_PUBLIC_HANKO_API!;

const HankoProfile = dynamic(() => import("../components/HankoProfile"), {
  ssr: false,
});

const Todo: NextPage = () => {
  const router = useRouter();
  const hankoClient = useMemo(() => createHankoClient(hankoAPI), []);

  const [error, setError] = useState<Error | null>(null);

  const logout = () => {
    hankoClient.user
      .logout()
      .catch((e) => {
        setError(e);
      });
  };

  const redirectToTodos = () => {
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
        <button onClick={redirectToTodos} className={styles.button}>
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
