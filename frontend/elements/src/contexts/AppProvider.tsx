import { JSXInternal } from "preact/src/jsx";
import { ComponentChildren, createContext, h } from "preact";
import { TranslateProvider } from "@denysvuika/preact-translate";

import {
  Fragment,
  StateUpdater,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "preact/compat";

import {
  Hanko,
  HankoError,
  TechnicalError,
  State,
  FlowName,
  FlowError,
  LastLogin,
  StateInitConfig,
} from "@teamhanko/hanko-frontend-sdk";

import { Translations } from "../i18n/translations";

import Container from "../components/wrapper/Container";
import InitPage from "../pages/InitPage";
import LoginInitPage from "../pages/LoginInitPage";
import PasscodePage from "../pages/PasscodePage";
import RegisterPasskeyPage from "../pages/RegisterPasskeyPage";
import LoginPasswordPage from "../pages/LoginPasswordPage";
import EditPasswordPage from "../pages/EditPasswordPage";
import LoginMethodChooserPage from "../pages/LoginMethodChooser";
import RegistrationInitPage from "../pages/RegistrationInitPage";
import CreatePasswordPage from "../pages/CreatePasswordPage";
import ProfilePage from "../pages/ProfilePage";
import ErrorPage from "../pages/ErrorPage";
import CreateEmailPage from "../pages/CreateEmailPage";
import CreateUsernamePage from "../pages/CreateUsernamePage";
import CredentialOnboardingChooserPage from "../pages/CredentialOnboardingChooser";
import LoginOTPPage from "../pages/LoginOTPPage";
import LoginSecurityKeyPage from "../pages/LoginSecurityKeyPage";
import MFAMethodChooserPage from "../pages/MFAMethodChooserPage";
import CreateOTPSecretPage from "../pages/CreateOTPSecretPage";
import CreateSecurityKeyPage from "../pages/CreateSecurityKeyPage";
import DeviceTrustPage from "../pages/DeviceTrustPage";

import SignalLike = JSXInternal.SignalLike;

export type ComponentName =
  | "auth"
  | "login"
  | "registration"
  | "profile"
  | "events";

export type HankoAuthMode = "registration" | "login";

export interface GlobalOptions {
  hanko?: Hanko;
  injectStyles?: boolean;
  enablePasskeys?: boolean;
  hidePasskeyButtonOnLogin?: boolean;
  translations?: Translations;
  translationsLocation?: string;
  fallbackLanguage?: string;
  storageKey?: string;
}

interface UIState {
  username?: string;
  email?: string;
  error?: FlowError;
  isDisabled?: boolean;
}

interface Context {
  hanko: Hanko;
  setHanko: StateUpdater<Hanko>;
  page: h.JSX.Element;
  setPage: StateUpdater<h.JSX.Element>;
  init: (compName: ComponentName) => void;
  componentName: ComponentName;
  setComponentName: StateUpdater<ComponentName>;
  lang: string;
  hidePasskeyButtonOnLogin: boolean;
  prefilledEmail?: string;
  prefilledUsername?: string;
  uiState: UIState;
  setUIState: StateUpdater<UIState>;
  initialComponentName: ComponentName;
  lastLogin?: LastLogin;
  isOwnFlow: (state: State<any>) => boolean;
}

export const AppContext = createContext<Context>(null);

interface Props {
  lang?: string | SignalLike<string>;
  prefilledEmail?: string;
  prefilledUsername?: string;
  mode?: HankoAuthMode;
  nonce?: string;
  componentName: ComponentName;
  globalOptions: GlobalOptions;
  children?: ComponentChildren;
  createWebauthnAbortSignal: () => AbortSignal;
}

const AppProvider = ({
  lang,
  prefilledEmail,
  prefilledUsername,
  globalOptions,
  createWebauthnAbortSignal,
  nonce,
  ...props
}: Props) => {
  const {
    hanko,
    injectStyles,
    hidePasskeyButtonOnLogin,
    translations,
    translationsLocation,
    fallbackLanguage,
  } = globalOptions;

  // Without this, the initial "lang" attribute value sometimes appears to not
  // be set properly. This results in a wrong X-Language header value being sent
  // to the API and hence in outgoing emails translated in the wrong language.
  hanko.setLang(lang?.toString() || fallbackLanguage);

  const ref = useRef<HTMLElement>(null);

  const storageKeyLastLogin = useMemo(
    () => `${globalOptions.storageKey}_last_login`,
    [globalOptions.storageKey],
  );

  const [componentName, setComponentName] = useState<ComponentName>(
    props.componentName,
  );

  const [authComponentFlow, setAuthComponentFlow] = useState<FlowName>(
    props.mode ?? "login",
  );

  // TODO: check if necessary, see commit e5e84de9 for more info
  const hasInitializedRef = useRef(false);
  const [isReadyToInit, setIsReadyToInit] = useState(false);

  const componentFlowNameMap = useMemo<Record<ComponentName, FlowName>>(
    () => ({
      auth: authComponentFlow,
      login: "login",
      registration: "registration",
      profile: "profile",
      events: null,
    }),
    [authComponentFlow],
  );

  const initComponent = useMemo(() => <InitPage />, []);
  const [page, setPage] = useState<h.JSX.Element>(initComponent);
  const [, setHanko] = useState<Hanko>(hanko);
  const [lastLogin, setLastLogin] = useState<LastLogin>();
  const [uiState, setUIState] = useState<UIState>({
    email: prefilledEmail,
    username: prefilledUsername,
  });

  const dispatchEvent = function <T>(type: string, detail?: T) {
    ref.current?.dispatchEvent(
      new CustomEvent<T>(type, {
        detail,
        bubbles: false,
        composed: true,
      }),
    );
  };

  const isOwnFlow = useCallback(
    (state: State<any>) =>
      componentFlowNameMap[componentName] == state.flowName,
    [componentFlowNameMap, componentName, authComponentFlow],
  );

  const handleError = (e: any) => {
    setPage(
      <ErrorPage error={e instanceof HankoError ? e : new TechnicalError(e)} />,
    );
  };

  useMemo(
    () =>
      hanko.onBeforeStateChange(({ state }) => {
        if (!isOwnFlow(state)) {
          return;
        }

        setUIState((prev) => ({ ...prev, isDisabled: true, error: undefined }));
      }),
    [hanko, isOwnFlow],
  );

  useEffect(() => {
    setUIState((prev) => ({
      ...prev,
      ...(prefilledEmail && { email: prefilledEmail }),
      ...(prefilledUsername && { username: prefilledUsername }),
    }));
  }, [prefilledEmail, prefilledUsername]);

  useEffect(
    () =>
      hanko.onAfterStateChange(async ({ state }) => {
        if (!isOwnFlow(state)) {
          return;
        }
        if (
          ![
            "onboarding_verify_passkey_attestation",
            "webauthn_credential_verification",
            "login_passkey",
            "thirdparty",
          ].includes(state.name)
        ) {
          setUIState((prev) => ({ ...prev, isDisabled: false }));
        }

        switch (state.name) {
          case "login_init":
            setPage(<LoginInitPage state={state} />);
            state.passkeyAutofillActivation();
            break;
          case "passcode_confirmation":
            setPage(<PasscodePage state={state} />);
            break;
          case "login_otp":
            setPage(<LoginOTPPage state={state} />);
            break;
          case "onboarding_create_passkey":
            setPage(<RegisterPasskeyPage state={state} />);
            break;
          case "login_password":
            setPage(<LoginPasswordPage state={state} />);
            break;
          case "login_password_recovery":
            setPage(<EditPasswordPage state={state} />);
            break;
          case "login_security_key":
            setPage(<LoginSecurityKeyPage state={state} />);
            break;
          case "mfa_method_chooser":
            setPage(<MFAMethodChooserPage state={state} />);
            break;
          case "mfa_otp_secret_creation":
            setPage(<CreateOTPSecretPage state={state} />);
            break;
          case "mfa_security_key_creation":
            setPage(<CreateSecurityKeyPage state={state} />);
            break;
          case "login_method_chooser":
            setPage(<LoginMethodChooserPage state={state} />);
            break;
          case "registration_init":
            setPage(<RegistrationInitPage state={state} />);
            break;
          case "password_creation":
            setPage(<CreatePasswordPage state={state} />);
            break;
          case "success":
            if (state.payload?.last_login) {
              localStorage.setItem(
                storageKeyLastLogin,
                JSON.stringify(state.payload.last_login),
              );
            }
            state.autoStep();
            break;
          case "profile_init":
            setPage(
              <ProfilePage
                state={state}
                enablePasskeys={globalOptions.enablePasskeys}
              />,
            );
            break;
          case "error":
            setPage(<ErrorPage state={state} />);
            break;
          case "onboarding_email":
            setPage(<CreateEmailPage state={state} />);
            break;
          case "onboarding_username":
            setPage(<CreateUsernamePage state={state} />);
            break;
          case "credential_onboarding_chooser":
            setPage(<CredentialOnboardingChooserPage state={state} />);
            break;
          case "device_trust":
            setPage(<DeviceTrustPage state={state} />);
            break;
        }
      }),
    [componentName, componentFlowNameMap],
  );

  const flowInit = useCallback(async (flowName: FlowName) => {
    setUIState((prev) => ({ ...prev, isDisabled: true }));
    const lastLoginEncoded = localStorage.getItem(storageKeyLastLogin);
    if (lastLoginEncoded) {
      setLastLogin(JSON.parse(lastLoginEncoded) as LastLogin);
    }
    const samlHint = new URLSearchParams(window.location.search).get(
      "saml_hint",
    );
    const config: StateInitConfig = {
      excludeAutoSteps: ["success"],
      cacheKey: `hanko-auth-flow-state`,
      dispatchAfterStateChangeEvent: false,
    };

    if (samlHint === "idp_initiated") {
      setAuthComponentFlow("token_exchange");
      await hanko.createState("token_exchange", {
        ...config,
        dispatchAfterStateChangeEvent: true,
      });
    } else {
      const state = await hanko.createState(flowName, config);
      setAuthComponentFlow(state.flowName);
      setTimeout(() => state.dispatchAfterStateChangeEvent(), 500);
    }
  }, []);

  const init = useCallback(
    (compName: ComponentName) => {
      setComponentName(compName);
      const flowName = componentFlowNameMap[compName];

      if (flowName) {
        flowInit(flowName).catch(handleError);
      }
    },
    [componentFlowNameMap],
  );

  // TODO: check if this can be done in cleaner way, see commit e5e84de9 for more info.
  // Step 1: Set the authComponentFlow from props.mode
  useEffect(() => {
    if (!hasInitializedRef.current) {
      const timer = setTimeout(() => {
        setAuthComponentFlow(props.mode ?? "login");
        setIsReadyToInit(true);
      }, 0);

      return () => clearTimeout(timer);
    }
  }, [props.mode]);

  // Step 2: Call init after authComponentFlow has been updated
  useEffect(() => {
    if (isReadyToInit && !hasInitializedRef.current) {
      hasInitializedRef.current = true;
      init(componentName);
    }
  }, [isReadyToInit, authComponentFlow, componentName, init]);

  useEffect(() => {
    hanko.onUserDeleted(() => {
      dispatchEvent("onUserDeleted");
    });

    hanko.onSessionCreated((detail) => {
      dispatchEvent("onSessionCreated", detail);
    });

    hanko.onSessionExpired(() => {
      dispatchEvent("onSessionExpired");
    });

    hanko.onUserLoggedOut(() => {
      dispatchEvent("onUserLoggedOut");
    });

    hanko.onBeforeStateChange((detail) => {
      dispatchEvent("onBeforeStateChange", detail);
    });

    hanko.onAfterStateChange((detail) => {
      dispatchEvent("onAfterStateChange", detail);
    });
  }, [hanko]);

  useMemo(() => {
    const cb = () => {
      init(componentName);
    };
    if (["auth", "login", "registration"].includes(componentName)) {
      hanko.onUserLoggedOut(cb);
      hanko.onSessionExpired(cb);
      hanko.onUserDeleted(cb);
    } else if (componentName === "profile") {
      hanko.onSessionCreated(cb);
    }
  }, [componentName, hanko, init]);

  return (
    <AppContext.Provider
      value={{
        init,
        initialComponentName: props.componentName,
        setUIState,
        uiState,
        hanko,
        setHanko,
        lang: lang?.toString() || fallbackLanguage,
        prefilledEmail,
        prefilledUsername,
        componentName,
        setComponentName,
        hidePasskeyButtonOnLogin,
        page,
        setPage,
        lastLogin,
        isOwnFlow,
      }}
    >
      <TranslateProvider
        translations={translations}
        fallbackLang={fallbackLanguage}
        root={translationsLocation}
      >
        <Container ref={ref}>
          {componentName !== "events" ? (
            <Fragment>
              {injectStyles ? (
                <style
                  nonce={nonce || undefined}
                  /* eslint-disable-next-line react/no-danger */
                  dangerouslySetInnerHTML={{
                    __html: window._hankoStyle.innerHTML,
                  }}
                />
              ) : null}
              {page}
            </Fragment>
          ) : null}
        </Container>
      </TranslateProvider>
    </AppContext.Provider>
  );
};

export default AppProvider;
