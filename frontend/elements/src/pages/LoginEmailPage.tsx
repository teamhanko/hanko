import {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "preact/compat";

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
  WebauthnFinalized,
} from "@teamhanko/hanko-frontend-sdk";

import { AppContext } from "../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

import Button from "../components/form/Button";
import Input from "../components/form/Input";
import Content from "../components/wrapper/Content";
import Form from "../components/form/Form";
import Divider from "../components/spacer/Divider";
import ErrorMessage from "../components/error/ErrorMessage";
import Headline1 from "../components/headline/Headline1";
import { IconName } from "../components/icons/Icon";

import LoginPasscodePage from "./LoginPasscodePage";
import RegisterConfirmPage from "./RegisterConfirmPage";
import LoginPasswordPage from "./LoginPasswordPage";
import RegisterPasskeyPage from "./RegisterPasskeyPage";
import RegisterPasswordPage from "./RegisterPasswordPage";
import ErrorPage from "./ErrorPage";
import AccountNotFoundPage from "./AccountNotFoundPage";

interface Props {
  emailAddress?: string;
}

const LoginEmailPage = (props: Props) => {
  const { t } = useContext(TranslateContext);
  const {
    hanko,
    prefilledEmail,
    experimentalFeatures,
    emitSuccessEvent,
    enablePasskeys,
    hidePasskeyButtonOnLogin,
    config,
    setPage,
    setPasscode,
    setUserInfo,
    setUser,
  } = useContext(AppContext);

  const [emailAddress, setEmailAddress] = useState<string>(
    props.emailAddress || prefilledEmail || ""
  );
  const [isPasskeyLoginLoading, setIsPasskeyLoginLoading] = useState<boolean>();
  const [isPasskeyLoginSuccess, setIsPasskeyLoginSuccess] = useState<boolean>();
  const [isEmailLoginLoading, setIsEmailLoginLoading] = useState<boolean>();
  const [error, setError] = useState<HankoError>(null);
  const [isConditionalMediationSupported, setIsConditionalMediationSupported] =
    useState<boolean>();
  const [isEmailLoginSuccess, setIsEmailLoginSuccess] = useState<boolean>();
  const [isThirdPartyLoginLoading, setIsThirdPartyLoginLoading] =
    useState<string>("");

  const isWebAuthnSupported = WebauthnSupport.supported();

  const disabled = useMemo(
    () =>
      isEmailLoginLoading ||
      isEmailLoginSuccess ||
      isPasskeyLoginLoading ||
      isPasskeyLoginSuccess ||
      !!isThirdPartyLoginLoading,
    [
      isEmailLoginLoading,
      isEmailLoginSuccess,
      isPasskeyLoginLoading,
      isPasskeyLoginSuccess,
      isThirdPartyLoginLoading,
    ]
  );

  const onEmailInput = (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      setEmailAddress(event.target.value);
    }
  };

  const onThirdPartyAuth = (event: Event, provider: string) => {
    event.preventDefault();
    setIsThirdPartyLoginLoading(provider);
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
            if (shouldRegisterPasskey && enablePasskeys) {
              setPage(<RegisterPasskeyPage />);
              return;
            }
            emitSuccessEvent(_user.id);
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
    [
      emitSuccessEvent,
      enablePasskeys,
      hanko.user,
      hanko.webauthn,
      setPage,
      setUser,
    ]
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

  const renderAccountNotFound = useCallback(
    () => setPage(<AccountNotFoundPage emailAddress={emailAddress} onBack={onBackHandler}/>), 
    [ 
      emailAddress, 
      onBackHandler, 
      setPage
    ]
  );

  const loginWithEmailAndWebAuthn = () => {
    let _userInfo: UserInfo;
    let _webauthnFinalizedResponse: WebauthnFinalized;
    let webauthnLoginInitiated: boolean;

    return hanko.user
      .getInfo(emailAddress)
      .then((resp) => setUserInfo((_userInfo = resp)))
      .then((): Promise<void | WebauthnFinalized> => {
        if (!_userInfo.verified && config.emails.require_verification) {
          return renderPasscode(_userInfo.id, _userInfo.email_id);
        }

        if (!_userInfo.has_webauthn_credential || conditionalMediationEnabled) {
          return renderAlternateLoginMethod(_userInfo);
        }

        webauthnLoginInitiated = true;
        return hanko.webauthn.login(_userInfo.id);
      })
      .then((resp: void | WebauthnFinalized) => {
        if (resp instanceof Object) {
          _webauthnFinalizedResponse = resp;
        }
        return;
      })
      .then(() => {
        if (webauthnLoginInitiated) {
          setError(null);
          setIsEmailLoginLoading(false);
          setIsEmailLoginSuccess(true);
          emitSuccessEvent(_webauthnFinalizedResponse.user_id);
        }

        return;
      })
      .catch((e) => {
        if (e instanceof NotFoundError) {
          
          if (config.account.allow_signup) {
            renderRegistrationConfirm();
            return;
          }
          
          renderAccountNotFound();
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

    if (isWebAuthnSupported && enablePasskeys) {
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
      .then((resp) => {
        setError(null);
        setIsPasskeyLoginLoading(false);
        setIsPasskeyLoginSuccess(true);
        emitSuccessEvent(resp.user_id);

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
      .then((resp) => {
        setError(null);
        emitSuccessEvent(resp.user_id);
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
  }, [conditionalMediationEnabled, emitSuccessEvent, hanko]);

  useEffect(() => {
    loginViaConditionalUI();
  }, [loginViaConditionalUI]);

  useEffect(() => {
    WebauthnSupport.isConditionalMediationAvailable()
      .then((supported) => setIsConditionalMediationSupported(supported))
      .catch((e) => setError(new TechnicalError(e)));
  }, []);

  useEffect(() => {
    if (isThirdPartyLoginLoading) {
      hanko.thirdParty
        .auth(isThirdPartyLoginLoading, window.location.href)
        .catch((error) => {
          setPage(<ErrorPage initialError={error} />);
        });
    }
  }, [hanko, setPage, isThirdPartyLoginLoading]);

  useEffect(() => {
    if (emailAddress.length === 0 && prefilledEmail !== undefined) {
      setEmailAddress(prefilledEmail);
    }
    // The dependency array is missing the emailAddress parameter intentionally because if it is not missing the email
    // would always be reset to the prefilledEmail when the input is empty
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [prefilledEmail]);

  return (
    <Content>
      <Headline1>{config.account.allow_signup ? t("headlines.loginEmail") : t("headlines.loginEmailNoSignup")}</Headline1>
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
        />
        <Button
          isLoading={isEmailLoginLoading}
          isSuccess={isEmailLoginSuccess}
          disabled={disabled}
        >
          {t("labels.continue")}
        </Button>
      </Form>
      {(enablePasskeys && !hidePasskeyButtonOnLogin) ||
      config.providers?.length ? (
        <Divider>{t("labels.or")}</Divider>
      ) : null}
      {enablePasskeys && !hidePasskeyButtonOnLogin ? (
        <Form onSubmit={onPasskeySubmit}>
          <Button
            secondary
            title={
              !isWebAuthnSupported ? t("labels.webauthnUnsupported") : null
            }
            isLoading={isPasskeyLoginLoading}
            isSuccess={isPasskeyLoginSuccess}
            disabled={!isWebAuthnSupported || disabled}
            icon={"passkey"}
          >
            {t("labels.signInPasskey")}
          </Button>
        </Form>
      ) : null}
      {config.providers?.map((provider: string) => (
        <Form
          key={provider}
          onSubmit={(e) => {
            onThirdPartyAuth(e, provider);
          }}
        >
          <Button
            secondary
            isLoading={isThirdPartyLoginLoading === provider}
            disabled={disabled}
            icon={provider.toLowerCase() as IconName}
          >
            {t("labels.signInWith", {
              provider,
            })}
          </Button>
        </Form>
      ))}
    </Content>
  );
};

export default LoginEmailPage;
