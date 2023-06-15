import { useCallback, useContext, useEffect } from "preact/compat";

import {
  HankoError,
  UnauthorizedError,
  User,
} from "@teamhanko/hanko-frontend-sdk";

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
    enablePasskeys,
    setConfig,
    setUser,
    setEmails,
    setWebauthnCredentials,
    setPage,
  } = useContext(AppContext);

  const afterLogin = useCallback(() => {
    let _user: User;
    return Promise.all([
      hanko.config.get().then(setConfig),
      hanko.user.getCurrent().then((resp) => setUser((_user = resp))),
    ])
      .then(() => hanko.webauthn.shouldRegister(_user))
      .then((shouldRegister) =>
        setPage(
          shouldRegister && enablePasskeys ? (
            <RegisterPasskeyPage />
          ) : (
            <LoginFinishedPage />
          )
        )
      );
  }, [enablePasskeys, hanko, setConfig, setPage, setUser]);

  const hankoAuthInit = useCallback(() => {
    const thirdPartyError = hanko.thirdParty.getError();

    if (thirdPartyError) {
      window.history.replaceState(null, null, window.location.pathname);
      setPage(<ErrorPage initialError={thirdPartyError} />);
      return;
    }

    const params = new URLSearchParams(window.location.search);
    const token = params.get("hanko_token");

    if (token && token.length) {
      hanko.token
        .validate()
        .then(() => afterLogin())
        .catch((e) => setPage(<ErrorPage initialError={e} />));
    } else {
      hanko.config
        .get()
        .then(setConfig)
        .then(() => setPage(<LoginEmailPage />))
        .catch((e) => setPage(<ErrorPage initialError={e} />));
    }
  }, [afterLogin, hanko, setConfig, setPage]);

  const initHankoProfile = useCallback(() => {
    Promise.all([
      hanko.config.get().then(setConfig),
      hanko.user.getCurrent().then(setUser),
      hanko.email.list().then(setEmails),
      hanko.webauthn.listCredentials().then(setWebauthnCredentials),
    ])
      .then(() => setPage(<ProfilePage />))
      .catch((e) => setPage(<ErrorPage initialError={e} />));
  }, [hanko, setConfig, setEmails, setPage, setUser, setWebauthnCredentials]);

  const showErrorPage = useCallback(
    (e: HankoError) => {
      setPage(<ErrorPage initialError={e} />);
    },
    [setPage]
  );

  useEffect(() => {
    if (componentName !== "auth") return;
    return hanko.onSessionExpired(() => hankoAuthInit());
  }, [componentName, hanko, hankoAuthInit]);

  useEffect(() => {
    if (componentName !== "profile") return;
    return hanko.onSessionCreated(() => initHankoProfile());
  }, [componentName, hanko, initHankoProfile]);

  useEffect(() => {
    const sessionIsValid = hanko.session.isValid();

    switch (componentName) {
      case "auth":
        if (sessionIsValid) {
          afterLogin().catch(showErrorPage);
        } else {
          hankoAuthInit();
        }
        break;
      case "profile":
        if (sessionIsValid) {
          initHankoProfile();
        } else {
          showErrorPage(new UnauthorizedError());
        }
        break;
    }
  }, [
    afterLogin,
    componentName,
    hanko,
    hankoAuthInit,
    initHankoProfile,
    setPage,
    showErrorPage,
  ]);

  return <LoadingSpinner isLoading />;
};

export default InitPage;
