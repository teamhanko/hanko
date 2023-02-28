import type { NextPage } from "next";
import dynamic from "next/dynamic";
import styles from "../styles/Todo.module.css";
import React, { useState } from "react";

const HankoAuth = dynamic(() => import("../components/HankoAuth"), {
  ssr: false,
});

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
