"use client";

import { useState } from "react";
import HankoAuth from "../components/HankoAuth";
import styles from "../styles/Todo.module.css";

export default function Home() {
  const [error, setError] = useState<Error | null>(null);

  return (
    <div className={styles.content}>
      <div className={styles.error}>{error?.message}</div>
      <HankoAuth setError={setError} />
    </div>
  );
}
