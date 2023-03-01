import * as preact from "preact";
import { useCallback, useContext, useEffect } from "preact/compat";

import { User } from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../contexts/AppProvider";

import ErrorPage from "./ErrorPage";
import ProfilePage from "./ProfilePage";
import LoginEmailPage from "./LoginEmailPage";
import LoginFinishedPage from "./LoginFinishedPage";
import RegisterPasskeyPage from "./RegisterPasskeyPage";

import LoadingSpinner from "../components/icons/LoadingSpinner";

const InitPage = () => {
  const {
    hanko,
    componentName,
    setConfig,
    setUser,
    setEmails,
    setWebauthnCredentials,
    setPage,
  } = useContext(AppContext);

  const afterLogin = useCallback(
    (_user: User) =>
      hanko.webauthn
        .shouldRegister(_user)
        .then((shouldRegister) =>
          shouldRegister ? <RegisterPasskeyPage /> : <LoginFinishedPage />
        ),
    [hanko.webauthn]
  );

  const initHankoAuth = useCallback(() => {
    let _user: User;
    return Promise.allSettled([
      hanko.config.get().then(setConfig),
      hanko.user.getCurrent().then((resp) => setUser((_user = resp))),
    ]).then(([configResult, userResult]) => {
      if (configResult.status === "rejected") {
        return <ErrorPage initialError={configResult.reason} />;
      }
      if (userResult.status === "fulfilled") {
        return afterLogin(_user);
      }
      return <LoginEmailPage />;
    });
  }, [afterLogin, hanko.config, hanko.user, setConfig, setUser]);

  const initHankoProfile = useCallback(
    () =>
      Promise.all([
        hanko.config.get().then(setConfig),
        hanko.user.getCurrent().then(setUser),
        hanko.email.list().then(setEmails),
        hanko.webauthn.listCredentials().then(setWebauthnCredentials),
      ]).then(() => <ProfilePage />),
    [hanko, setConfig, setEmails, setUser, setWebauthnCredentials]
  );

  const getInitializer = useCallback(() => {
    switch (componentName) {
      case "auth":
        return initHankoAuth;
      case "profile":
        return initHankoProfile;
      default:
        return;
    }
  }, [componentName, initHankoAuth, initHankoProfile]);

  useEffect(() => {
    const initializer = getInitializer();
    if (initializer) {
      initializer()
        .then(setPage)
        .catch((e) => setPage(<ErrorPage initialError={e} />));
    }
  }, [getInitializer, setPage]);

  return <LoadingSpinner isLoading />;
};

export default InitPage;
