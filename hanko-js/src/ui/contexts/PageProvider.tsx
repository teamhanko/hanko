import * as preact from "preact";
import { createContext, h } from "preact";
import { useCallback, useContext, useMemo, useState } from "preact/compat";
import { HankoError } from "../../lib/Errors";

import { User } from "../../lib/HankoClient";

import { AppContext } from "./AppProvider";
import { PasswordContext } from "./PasswordProvider";
import { PasscodeContext } from "./PasscodeProvider";

import Initialize from "../pages/Initialize";
import LoginEmail from "./../pages/LoginEmail";
import LoginPasscode from "./../pages/LoginPasscode";
import LoginPassword from "./../pages/LoginPassword";
import LoginFinished from "./../pages/LoginFinished";
import RegisterConfirm from "./../pages/RegisterConfirm";
import RegisterPassword from "./../pages/RegisterPassword";
import RegisterAuthenticator from "./../pages/RegisterAuthenticator";
import Error from "./../pages/Error";
import Container from "../components/Container";

interface Context {
  emitSuccessEvent: () => void;
  eventuallyRenderEnrollment: (
    user: User,
    recoverPassword: boolean
  ) => Promise<boolean>;
  renderPassword: (userID: string) => Promise<void>;
  renderPasscode: (
    userID: string,
    recoverPassword: boolean,
    hideBackButton: boolean
  ) => Promise<void>;
  renderError: (e: HankoError) => void;
  renderLoginEmail: () => void;
  renderLoginFinished: () => void;
  renderRegisterConfirm: () => void;
  renderRegisterAuthenticator: () => void;
}

export const RenderContext = createContext<Context>(null);

const PageProvider = () => {
  const { hanko } = useContext(AppContext);
  const { passwordInitialize } = useContext(PasswordContext);
  const { passcodeInitialize } = useContext(PasscodeContext);

  const [page, setPage] = useState<h.JSX.Element>(<Initialize />);
  const [loginFinished, setLoginFinished] = useState<boolean>(false);

  const emitSuccessEvent = useCallback(() => {
    setLoginFinished(true);
  }, []);

  const pages = useMemo(
    () => ({
      loginEmail: () => setPage(<LoginEmail />),
      loginPasscode: (
        userID: string,
        recoverPassword: boolean,
        initialError?: HankoError,
        hideBackLink?: boolean
      ) =>
        setPage(
          <LoginPasscode
            userID={userID}
            recoverPassword={recoverPassword}
            initialError={initialError}
            hideBackLink={hideBackLink}
          />
        ),
      loginPassword: (userID: string, initialError: HankoError) =>
        setPage(<LoginPassword userID={userID} initialError={initialError} />),
      registerConfirm: () => setPage(<RegisterConfirm />),
      registerPassword: (user: User, enrollWebauthn: boolean) =>
        setPage(
          <RegisterPassword
            user={user}
            registerAuthenticator={enrollWebauthn}
          />
        ),
      registerAuthenticator: () => setPage(<RegisterAuthenticator />),
      loginFinished: () => setPage(<LoginFinished />),
      error: (error: HankoError) => setPage(<Error initialError={error} />),
    }),
    []
  );

  const renderLoginEmail = useCallback(() => {
    pages.loginEmail();
  }, [pages]);

  const renderLoginFinished = useCallback(() => {
    pages.loginFinished();
  }, [pages]);

  const renderPassword = useCallback(
    (userID: string) => {
      return new Promise<void>((resolve, reject) => {
        passwordInitialize(userID)
          .then((e) => pages.loginPassword(userID, e))
          .catch((e) => reject(e));
      });
    },
    [pages, passwordInitialize]
  );

  const renderPasscode = useCallback(
    (userID: string, recoverPassword: boolean, hideBackButton: boolean) => {
      return new Promise<void>((resolve, reject) => {
        passcodeInitialize(userID)
          .then((e) => {
            pages.loginPasscode(userID, recoverPassword, e, hideBackButton);

            return resolve();
          })
          .catch((e) => reject(e));
      });
    },
    [pages, passcodeInitialize]
  );

  const eventuallyRenderEnrollment = useCallback(
    (user: User, recoverPassword: boolean) => {
      return new Promise<boolean>((resolve, reject) => {
        hanko.authenticator
          .shouldRegister(user)
          .then((shouldRegisterAuthenticator) => {
            let rendered = true;
            if (recoverPassword) {
              pages.registerPassword(user, shouldRegisterAuthenticator);
            } else if (shouldRegisterAuthenticator) {
              pages.registerAuthenticator();
            } else {
              rendered = false;
            }
            return resolve(rendered);
          })
          .catch((e) => reject(e));
      });
    },
    [hanko, pages]
  );

  const renderRegisterConfirm = useCallback(() => {
    pages.registerConfirm();
  }, [pages]);

  const renderRegisterAuthenticator = useCallback(() => {
    pages.registerAuthenticator();
  }, [pages]);

  const renderError = useCallback(
    (e: HankoError) => {
      pages.error(e);
    },
    [pages]
  );

  return (
    <RenderContext.Provider
      value={{
        emitSuccessEvent,
        renderLoginEmail,
        renderLoginFinished,
        renderPassword,
        renderPasscode,
        eventuallyRenderEnrollment,
        renderRegisterConfirm,
        renderRegisterAuthenticator,
        renderError,
      }}
    >
      <Container emitSuccessEvent={loginFinished}>{page}</Container>
    </RenderContext.Provider>
  );
};

export default PageProvider;
