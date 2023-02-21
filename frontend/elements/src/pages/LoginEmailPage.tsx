import * as preact from "preact";
import {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "preact/compat";
import { Fragment } from "preact";

import {
  HankoError,
  TechnicalError,
  NotFoundError,
  WebauthnRequestCancelledError,
  InvalidWebauthnCredentialError,
  TooManyRequestsError,
  WebauthnSupport,
  UserInfo,
  User,
} from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Button from "../components/form/Button";
import Input from "../components/form/Input";
import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Divider from "../components/divider/Divider";
import ErrorMessage from "../components/error/ErrorMessage";
import Headline1 from "../components/headline/Headline1";

import LoginPasscodePage from "./LoginPasscodePage";
import RegisterConfirmPage from "./RegisterConfirmPage";
import LoginPasswordPage from "./LoginPasswordPage";
import RegisterPasskeyPage from "./RegisterPasskeyPage";
import RegisterPasswordPage from "./RegisterPasswordPage";
import ErrorPage from "./ErrorPage";

interface Props {
  emailAddress?: string;
}

const LoginEmailPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const {
    hanko,
    experimentalFeatures,
    emitSuccessEvent,
    config,
    setPage,
    setPasscode,
    setUserInfo,
    setUser,
  } = useContext(AppContext);

  const [emailAddress, setEmailAddress] = useState<string>(props.emailAddress);
  const [isPasskeyLoginLoading, setIsPasskeyLoginLoading] = useState<boolean>();
  const [isPasskeyLoginSuccess, setIsPasskeyLoginSuccess] = useState<boolean>();
  const [isEmailLoginLoading, setIsEmailLoginLoading] = useState<boolean>();
  const [error, setError] = useState<HankoError>(null);
  const [isWebAuthnSupported, setIsWebAuthnSupported] = useState<boolean>();
  const [isConditionalMediationSupported, setIsConditionalMediationSupported] =
    useState<boolean>();
  const [isEmailLoginSuccess, setIsEmailLoginSuccess] = useState<boolean>();

  const disabled = useMemo(
    () =>
      isEmailLoginLoading ||
      isEmailLoginSuccess ||
      isPasskeyLoginLoading ||
      isPasskeyLoginSuccess,
    [
      isEmailLoginLoading,
      isEmailLoginSuccess,
      isPasskeyLoginLoading,
      isPasskeyLoginSuccess,
    ]
  );

  const onEmailInput = (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setEmailAddress(event.target.value);
    }
  };

  const onThirdPartyAuth = (event: Event, provider: string) => {
    event.preventDefault();
    hanko.thirdParty
      .auth(provider, window.location.href)
      .catch((error) => setPage(<ErrorPage initialError={error} />));
  };

  const onBackHandler = useCallback(() => {
    setPage(<LoginEmailPage emailAddress={emailAddress} />);
  }, [emailAddress, setPage]);

  const afterLoginHandler = useCallback(
    (recoverPassword: boolean) => {
      let _user: User;
      return hanko.user
        .getCurrent()
        .then((resp) => setUser((_user = resp)))
        .then(() => hanko.webauthn.shouldRegister(_user))
        .then((shouldRegisterPasskey) => {
          const onSuccessHandler = () => {
            if (shouldRegisterPasskey) {
              setPage(<RegisterPasskeyPage />);
              return;
            }
            emitSuccessEvent();
          };

          if (recoverPassword) {
            setPage(<RegisterPasswordPage onSuccess={onSuccessHandler} />);
          } else {
            onSuccessHandler();
          }

          return;
        })
        .catch((e) => setPage(<ErrorPage initialError={e} />));
    },
    [emitSuccessEvent, hanko.user, hanko.webauthn, setPage, setUser]
  );

  const renderPasscode = useCallback(
    (userID: string, emailID: string, recoverPassword?: boolean) => {
      const showPasscodePage = (e?: HankoError) =>
        setPage(
          <LoginPasscodePage
            userID={userID}
            emailID={emailID}
            emailAddress={emailAddress}
            initialError={e}
            onSuccess={() => afterLoginHandler(recoverPassword)}
            onBack={onBackHandler}
          />
        );

      return hanko.passcode
        .initialize(userID, emailID, false)
        .then(setPasscode)
        .then(() => showPasscodePage())
        .catch((e) => {
          if (e instanceof TooManyRequestsError) {
            showPasscodePage(e);
            return;
          }

          throw e;
        });
    },
    [
      afterLoginHandler,
      emailAddress,
      hanko.passcode,
      onBackHandler,
      setPage,
      setPasscode,
    ]
  );

  const renderRegistrationConfirm = useCallback(
    () =>
      setPage(
        <RegisterConfirmPage
          onSuccess={() => afterLoginHandler(config.password.enabled)}
          onPasscode={(userID: string, emailID: string) =>
            renderPasscode(userID, emailID, config.password.enabled)
          }
          emailAddress={emailAddress}
          onBack={onBackHandler}
        />
      ),
    [
      afterLoginHandler,
      config.password.enabled,
      emailAddress,
      onBackHandler,
      renderPasscode,
      setPage,
    ]
  );

  const loginWithEmailAndWebAuthn = () => {
    let _userInfo: UserInfo;
    let webauthnLoginInitiated: boolean;

    return hanko.user
      .getInfo(emailAddress)
      .then((resp) => setUserInfo((_userInfo = resp)))
      .then(() => {
        if (!_userInfo.verified && config.emails.require_verification) {
          return renderPasscode(_userInfo.id, _userInfo.email_id);
        }

        if (!_userInfo.has_webauthn_credential || conditionalMediationEnabled) {
          return renderAlternateLoginMethod(_userInfo);
        }

        webauthnLoginInitiated = true;
        return hanko.webauthn.login(_userInfo.id);
      })
      .then(() => {
        if (webauthnLoginInitiated) {
          setIsEmailLoginLoading(false);
          setIsEmailLoginSuccess(true);
          emitSuccessEvent();
        }

        return;
      })
      .catch((e) => {
        if (e instanceof NotFoundError) {
          renderRegistrationConfirm();
          return;
        }

        if (e instanceof WebauthnRequestCancelledError) {
          return renderAlternateLoginMethod(_userInfo);
        }

        throw e;
      });
  };

  const loginWithEmail = () => {
    let _userInfo: UserInfo;
    return hanko.user
      .getInfo(emailAddress)
      .then((resp) => setUserInfo((_userInfo = resp)))
      .then(() => {
        if (!_userInfo.verified && config.emails.require_verification) {
          return renderPasscode(_userInfo.id, _userInfo.email_id);
        }

        return renderAlternateLoginMethod(_userInfo);
      })
      .catch((e) => {
        if (e instanceof NotFoundError) {
          renderRegistrationConfirm();
          return;
        }

        throw e;
      });
  };

  const onEmailSubmit = (event: Event) => {
    event.preventDefault();
    setIsEmailLoginLoading(true);

    if (isWebAuthnSupported) {
      loginWithEmailAndWebAuthn().catch((e) => {
        setIsEmailLoginLoading(false);
        setError(e);
      });
    } else {
      loginWithEmail().catch((e) => {
        setIsEmailLoginLoading(false);
        setError(e);
      });
    }
  };

  const onPasskeySubmit = (event: Event) => {
    event.preventDefault();
    setIsPasskeyLoginLoading(true);

    hanko.webauthn
      .login()
      .then(() => {
        setError(null);
        setIsPasskeyLoginLoading(false);
        setIsPasskeyLoginSuccess(true);
        emitSuccessEvent();

        return;
      })
      .catch((e) => {
        setIsPasskeyLoginLoading(false);
        setError(e instanceof WebauthnRequestCancelledError ? null : e);
      });
  };

  const conditionalMediationEnabled = useMemo(
    () =>
      experimentalFeatures.includes("conditionalMediation") &&
      isConditionalMediationSupported,
    [experimentalFeatures, isConditionalMediationSupported]
  );

  const renderAlternateLoginMethod = useCallback(
    (_userInfo: UserInfo) => {
      if (config.password.enabled) {
        setPage(
          <LoginPasswordPage
            userInfo={_userInfo}
            onSuccess={() => afterLoginHandler(false)}
            onRecovery={() =>
              renderPasscode(_userInfo.id, _userInfo.email_id, true)
            }
            onBack={onBackHandler}
          />
        );
        return;
      }

      return renderPasscode(_userInfo.id, _userInfo.email_id);
    },
    [
      afterLoginHandler,
      config.password.enabled,
      onBackHandler,
      renderPasscode,
      setPage,
    ]
  );

  const loginViaConditionalUI = useCallback(() => {
    if (!conditionalMediationEnabled) {
      // Browser doesn't support AutoFill-assisted requests or the experimental conditional mediation feature is not enabled.
      return;
    }

    hanko.webauthn
      .login(null, true)
      .then(() => {
        setError(null);
        emitSuccessEvent();
        setIsEmailLoginSuccess(true);

        return;
      })
      .catch((e) => {
        if (e instanceof InvalidWebauthnCredentialError) {
          // An invalid WebAuthn credential has been used. Retry the login procedure, so another credential can be
          // chosen by the user via conditional UI.
          loginViaConditionalUI();
        }
        setError(e instanceof WebauthnRequestCancelledError ? null : e);
      });
  }, [conditionalMediationEnabled, emitSuccessEvent, hanko.webauthn]);

  useEffect(() => {
    loginViaConditionalUI();
  }, [loginViaConditionalUI]);

  useEffect(() => {
    setIsWebAuthnSupported(WebauthnSupport.supported());
  }, []);

  useEffect(() => {
    WebauthnSupport.isConditionalMediationAvailable()
      .then((supported) => setIsConditionalMediationSupported(supported))
      .catch((e) => setError(new TechnicalError(e)));
  }, []);

  return (
    <Content>
      <Headline1>{t("headlines.loginEmail")}</Headline1>
      <ErrorMessage error={error} />
      <Form onSubmit={onEmailSubmit}>
        <Input
          name={"email"}
          type={"email"}
          autoComplete={"username webauthn"}
          autoCorrect={"off"}
          required={true}
          onInput={onEmailInput}
          value={emailAddress}
          placeholder={t("labels.email")}
          pattern={"^.*[^0-9]+$"}
          disabled={disabled}
          autoFocus
        />
        <Button
          isLoading={isEmailLoginLoading}
          isSuccess={isEmailLoginSuccess}
          disabled={disabled}
        >
          {t("labels.continue")}
        </Button>
      </Form>
      {isWebAuthnSupported && !conditionalMediationEnabled ? (
        <Fragment>
          <Divider />
          <Form onSubmit={onPasskeySubmit}>
            <Button
              secondary
              isLoading={isPasskeyLoginLoading}
              isSuccess={isPasskeyLoginSuccess}
              disabled={disabled}
            >
              {t("labels.signInPasskey")}
            </Button>
          </Form>
          {config.providers?.map((provider: string) => {
            return (
              <Form
                key={provider}
                onSubmit={(e) => {
                  onThirdPartyAuth(e, provider);
                }}
              >
                <Button secondary>
                  {t("labels.signInWith", {
                    provider,
                  })}
                </Button>
              </Form>
            );
          })}
        </Fragment>
      ) : null}
    </Content>
  );
};

export default LoginEmailPage;
