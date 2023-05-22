import { useCallback, useContext, useEffect } from "preact/compat";

import { UnauthorizedError, User } from "@teamhanko/hanko-frontend-sdk";

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

  const afterLogin = useCallback(() => {
    let _user: User;
    return Promise.all([
      hanko.config.get().then(setConfig),
      hanko.user.getCurrent().then((resp) => setUser((_user = resp))),
    ])
      .then(() => hanko.webauthn.shouldRegister(_user))
      .then((shouldRegister) =>
        shouldRegister ? <RegisterPasskeyPage /> : <LoginFinishedPage />
      );
  }, [hanko.config, hanko.user, hanko.webauthn, setConfig, setUser]);

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
  }, [
    afterLogin,
    hanko.config,
    hanko.thirdParty,
    hanko.token,
    setConfig,
    setPage,
  ]);

  const initHankoProfile = useCallback(() => {
    Promise.all([
      hanko.config.get().then(setConfig),
      hanko.user.getCurrent().then(setUser),
      hanko.email.list().then(setEmails),
      hanko.webauthn.listCredentials().then(setWebauthnCredentials),
    ])
      .then(() => setPage(<ProfilePage />))
      .catch((e) => setPage(<ErrorPage initialError={e} />));
  }, [
    hanko.config,
    hanko.email,
    hanko.user,
    hanko.webauthn,
    setConfig,
    setEmails,
    setPage,
    setUser,
    setWebauthnCredentials,
  ]);

  useEffect(() => {
    if (componentName !== "auth") return;
    return hanko.onSessionNotPresent(() => hankoAuthInit());
  }, [componentName, hanko, hankoAuthInit]);

  useEffect(() => {
    if (componentName !== "auth") return;
    return hanko.onSessionExpired(() => hankoAuthInit());
  }, [componentName, hanko, hankoAuthInit]);

  useEffect(() => {
    if (componentName !== "auth") return;
    return hanko.onSessionResumed(() => {
      afterLogin().catch((e) => setPage(<ErrorPage initialError={e} />));
    });
  }, [afterLogin, componentName, hanko, setPage]);

  useEffect(() => {
    if (componentName !== "profile") return;
    return hanko.onSessionResumed(() => initHankoProfile());
  }, [componentName, hanko, initHankoProfile]);

  useEffect(() => {
    if (componentName !== "profile") return;
    return hanko.onSessionCreated(() => initHankoProfile());
  }, [componentName, hanko, initHankoProfile]);

  useEffect(() => {
    if (componentName !== "profile") return;
    return hanko.onSessionNotPresent(() =>
      setPage(<ErrorPage initialError={new UnauthorizedError()} />)
    );
  }, [componentName, hanko, initHankoProfile, setPage]);

  useEffect(() => {
    hanko.relay.dispatchInitialEvents();
  }, [hanko.relay]);
  return <LoadingSpinner isLoading />;
};

export default InitPage;
