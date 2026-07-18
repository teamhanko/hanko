"use client";

import { useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";
import styles from "../../styles/Todo.module.css";

import { Hanko } from "@teamhanko/hanko-elements";
import HankoProfile from "../../components/HankoProfile";
import { SessionExpiredModal } from "../../components/SessionExpiredModal";

export default function Profile() {
  const router = useRouter();
  const [hankoClient, setHankoClient] = useState<Hanko>();
  const modalRef = useRef<HTMLDialogElement>(null);
  const [error, setError] = useState<Error | null>(null);
  const api = process.env.NEXT_PUBLIC_HANKO_API!;

  const logout = () => {
    hankoClient?.user.logout().catch((e) => {
      setError(e);
    });
  };

  const redirectToTodos = () => {
    router.push("/todo");
  };

  const redirectToLogin = useCallback(() => {
    router.push("/");
  }, [router]);

  useEffect(() => {
    if (!hankoClient) {
      return;
    }

    if (!hankoClient.session.isValid()) {
      redirectToLogin();
    }
  }, [hankoClient, redirectToLogin]);

  useEffect(() => {
    setHankoClient(new Hanko(api));
  }, [api]);

  useEffect(() => {
    if (hankoClient) {
      hankoClient.onUserLoggedOut(() => {
        redirectToLogin();
      });
    }
  }, [hankoClient, redirectToLogin]);

  useEffect(() => {
    if (hankoClient) {
      hankoClient.onSessionExpired(() => {
        modalRef.current?.showModal();
      });
    }
  }, [hankoClient]);

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
}
