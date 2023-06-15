import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { Hanko, register } from "@teamhanko/hanko-elements";
import styles from "./Todo.module.css";
import { useNavigate } from "react-router-dom";
import { SessionExpiredModal } from "./SessionExpiredModal";

const api = process.env.REACT_APP_HANKO_API!;

function HankoProfile() {
  const navigate = useNavigate();
  const hankoClient = useMemo(() => new Hanko(api), []);
  const [error, setError] = useState<Error | null>(null);
  const modalRef = useRef<HTMLDialogElement>(null);

  const logout = () => {
    hankoClient.user.logout().catch(setError);
  };

  const redirectToTodo = () => {
    navigate("/todo");
  };

  const redirectToLogin = useCallback(() => {
    navigate("/");
  }, [navigate]);

  useEffect(() => {
    register(api).catch(setError);
  }, []);

  useEffect(() => {
    if (!hankoClient.session.isValid()) {
      redirectToLogin();
    }
  }, [hankoClient, redirectToLogin]);


  useEffect(
    () => hankoClient.onUserLoggedOut(() => redirectToLogin()),
    [hankoClient, redirectToLogin]
  );

  useEffect(
    () => hankoClient.onSessionExpired(() => modalRef.current?.showModal()),
    [hankoClient]
  );

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
