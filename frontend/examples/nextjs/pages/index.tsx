import type { NextPage } from "next";
import styles from "../styles/Todo.module.css";
import React, { useState } from "react";
import HankoAuth from "../components/HankoAuth";


const Home: NextPage = () => {
  const [error, setError] = useState<Error | null>(null);

  return (
    <div className={styles.content}>
      <div className={styles.error}>{error?.message}</div>
      <HankoAuth setError={setError} />
    </div>
  );
};

export default Home;
