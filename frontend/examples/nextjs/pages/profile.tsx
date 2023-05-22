import React, { useCallback, useEffect, useRef, useState } from "react";
import { NextPage } from "next";
import { useRouter } from "next/router";
import dynamic from "next/dynamic";
import styles from "../styles/Todo.module.css";

import { SessionExpiredModal } from "../components/SessionExpiredModal";
import { Hanko } from "@teamhanko/hanko-elements";

const hankoAPI = process.env.NEXT_PUBLIC_HANKO_API!;
const HankoProfile = dynamic(() => import("../components/HankoProfile"), {
  ssr: false,
});

const Profile: NextPage = () => {
  const router = useRouter();
  const [hankoClient, setHankoClient] = useState<Hanko>();

  useEffect(() => {
    import("@teamhanko/hanko-elements").then(({ Hanko }) => setHankoClient(new Hanko(hankoAPI)));
  }, []);

  const modalRef = useRef<HTMLDialogElement>(null);
  const [error, setError] = useState<Error | null>(null);

  const logout = () => {
    hankoClient?.user
      .logout()
      .catch((e) => {
        setError(e);
      });
  };

  const redirectToTodos = () => {
    router.push("/todo").catch((e) => setError(e));
  };

  const redirectToLogin = useCallback(() => {
    router.push("/").catch(setError)
  }, [router]);

  useEffect(() => hankoClient?.onUserLoggedOut(() => {
    redirectToLogin();
  }), [hankoClient, redirectToLogin]);

  useEffect(() => hankoClient?.onSessionNotPresent(() => {
    redirectToLogin();
  }), [hankoClient, redirectToLogin]);

  useEffect(() => hankoClient?.onSessionExpired(() => {
    modalRef.current?.showModal();
  }), [hankoClient]);

  return (
    <>
      <SessionExpiredModal ref={modalRef} />
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

export default Profile;
