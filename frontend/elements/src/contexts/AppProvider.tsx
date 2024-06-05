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
  create as createWebauthnCredential,
  get as getWebauthnCredential,
} from "@github/webauthn-json";

import {
  Hanko,
  TechnicalError,
  UnauthorizedError,
  WebauthnSupport,
} from "@teamhanko/hanko-frontend-sdk";

import { Translations } from "../i18n/translations";

import {
  FlowPath,
  Handlers,
} from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/state-handling";

import { Error as FlowError } from "@teamhanko/hanko-frontend-sdk/dist/lib/flow-api/types/error";

import {
  PublicKeyCredentialWithAssertionJSON,
  PublicKeyCredentialWithAttestationJSON,
} from "@github/webauthn-json/src/webauthn-json/basic/json";

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
import SignalLike = JSXInternal.SignalLike;
import CreateEmailPage from "../pages/CreateEmailPage";

type ExperimentalFeature = "conditionalMediation";
type ExperimentalFeatures = ExperimentalFeature[];

export type ComponentName =
  | "auth"
  | "login"
  | "registration"
  | "profile"
  | "events";

export interface GlobalOptions {
  hanko?: Hanko;
  injectStyles?: boolean;
  enablePasskeys?: boolean;
  hidePasskeyButtonOnLogin?: boolean;
  translations?: Translations;
  translationsLocation?: string;
  fallbackLanguage?: string;
}

export type UIAction =
  | "email-submit"
  | "passkey-submit"
  | "passkey-rename"
  | "passkey-delete"
  | "passcode-resend"
  | "passcode-submit"
  | "password-submit"
  | "password-recovery"
  | "password-delete"
  | "choose-login-method"
  | "switch-flow"
  | "email-set-primary"
  | "email-delete"
  | "email-verify"
  | "username-set"
  | "skip"
  | "back"
  | "account_delete"
  | "retry";

interface UIState {
  username?: string;
  email?: string;
  loadingAction?: UIAction;
  succeededAction?: UIAction;
  lastAction?: UIAction;
  error?: FlowError;
}

interface Context {
  hanko: Hanko;
  page: h.JSX.Element;
  setPage: StateUpdater<h.JSX.Element>;
  init: (compName: ComponentName) => void;
  isDisabled: boolean;
  componentName: ComponentName;
  setComponentName: StateUpdater<ComponentName>;
  experimentalFeatures?: ExperimentalFeatures;
  emitSuccessEvent: (userID: string) => void;
  lang: string;
  hidePasskeyButtonOnLogin: boolean;
  prefilledEmail?: string;
  prefilledUsername?: string;
  stateHandler: Handlers;
  setLoadingAction: StateUpdater<UIAction>;
  setSucceededAction: StateUpdater<UIAction>;
  uiState: UIState;
  setUIState: StateUpdater<UIState>;
  initialComponentName: ComponentName;
}

export const AppContext = createContext<Context>(null);

interface Props {
  lang?: string | SignalLike<string>;
  experimental?: string;
  prefilledEmail?: string;
  prefilledUsername?: string;
  componentName: ComponentName;
  globalOptions: GlobalOptions;
  children?: ComponentChildren;
}

const AppProvider = ({
  lang,
  experimental = "",
  prefilledEmail,
  prefilledUsername,
  globalOptions,
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

  const ref = useRef<HTMLElement>(null);

  const [componentName, setComponentName] = useState<ComponentName>(
    props.componentName,
  );

  const experimentalFeatures = useMemo(
    () =>
      experimental
        .split(" ")
        .filter((feature) => feature.length)
        .map((feature) => feature as ExperimentalFeature),
    [experimental],
  );

  const initComponent = useMemo(() => <InitPage />, []);
  const [page, setPage] = useState<h.JSX.Element>(initComponent);
  const [uiState, setUIState] = useState<UIState>({
    email: prefilledEmail,
    username: prefilledUsername,
  });
  let abortController = new AbortController();

  const setLoadingAction = useCallback((loadingAction: UIAction) => {
    setUIState((prev) => ({
      ...prev,
      loadingAction,
      succeededAction: null,
      error: null,
      lastAction: loadingAction || prev.lastAction,
    }));
  }, []);

  const setSucceededAction = useCallback((succeededAction: UIAction) => {
    setUIState((prev) => ({
      ...prev,
      succeededAction,
      loadingAction: null,
    }));
  }, []);

  const lastActionSucceeded = useCallback(() => {
    setUIState((prev) => ({
      ...prev,
      succeededAction: prev.lastAction,
      loadingAction: null,
      error: null,
    }));
  }, []);

  const isDisabled = useMemo(
    () => !!uiState.loadingAction || !!uiState.succeededAction,
    [uiState],
  );

  const dispatchEvent = function <T>(type: string, detail?: T) {
    ref.current?.dispatchEvent(
      new CustomEvent<T>(type, {
        detail,
        bubbles: false,
        composed: true,
      }),
    );
  };

  const emitSuccessEvent = useCallback(
    (userID: string) => {
      const event = new Event("hankoAuthSuccess", {
        bubbles: true,
        composed: true,
      });
      const fn = setTimeout(() => {
        hanko.relay.dispatchAuthFlowCompletedEvent({ userID });
        ref.current.dispatchEvent(event);
      }, 500);

      return () => clearTimeout(fn);
    },
    [hanko],
  );

  const handleError = (e: any) => {
    setLoadingAction(null);
    setPage(
      <ErrorPage
        error={
          hanko.flow.client.response?.status === 401
            ? new UnauthorizedError(e)
            : new TechnicalError(e)
        }
      />,
    );
  };

  const _createAbortSignal = () => {
    if (abortController) {
      console.log("_createAbortSignal abort");
      abortController.abort();
    }

    abortController = new AbortController();
    return abortController.signal;
  };

  const stateHandler: Handlers & { onError: (e: any) => void } = useMemo(
    () => ({
      onError: (e) => {
        handleError(e);
      },
      async preflight(state) {
        const conditionalMediationAvailable =
          await WebauthnSupport.isConditionalMediationAvailable();

        const newState = await state.actions
          .register_client_capabilities({
            webauthn_available: isWebAuthnSupported,
            webauthn_conditional_mediation_available:
              conditionalMediationAvailable,
          })
          .run();
        return hanko.flow.run(newState, stateHandler);
      },
      async login_init(state) {
        setPage(<LoginInitPage state={state} />);
        void (async function () {
          if (state.payload.request_options) {
            let assertionResponse: PublicKeyCredentialWithAssertionJSON;

            try {
              assertionResponse = await getWebauthnCredential({
                publicKey: state.payload.request_options.publicKey,
                mediation: "conditional" as CredentialMediationRequirement,
                signal: _createAbortSignal(),
              });
            } catch (error) {
              // We do not need to handle the error, because this is a conditional request, which can fail silently
              return;
            }

            setLoadingAction("passkey-submit");
            const nextState = await state.actions
              .webauthn_verify_assertion_response({
                assertion_response: assertionResponse,
              })
              .run();

            setLoadingAction(null);
            stateHandler[nextState.name](nextState);
          }
        })();
      },
      passcode_confirmation(state) {
        setPage(<PasscodePage state={state} />);
      },
      async login_passkey(state) {
        let assertionResponse: PublicKeyCredentialWithAssertionJSON;
        setLoadingAction("passkey-submit");

        try {
          assertionResponse = await getWebauthnCredential({
            ...state.payload.request_options,
            signal: _createAbortSignal(),
          });
        } catch (error) {
          const prevState = await state.actions.back(null).run();
          setLoadingAction(null);
          return hanko.flow.run(prevState, stateHandler);
        }

        const nextState = await state.actions
          .webauthn_verify_assertion_response({
            assertion_response: assertionResponse,
          })
          .run();

        setLoadingAction(null);
        stateHandler[nextState.name](nextState);
      },
      onboarding_create_passkey(state) {
        setPage(<RegisterPasskeyPage state={state} />);
      },
      async onboarding_verify_passkey_attestation(state) {
        let attestationResponse: PublicKeyCredentialWithAttestationJSON;
        try {
          attestationResponse = await createWebauthnCredential({
            ...state.payload.creation_options,
            signal: _createAbortSignal(),
          });
        } catch (e) {
          const prevState = await state.actions.back(null).run();
          setLoadingAction(null);
          stateHandler[prevState.name](prevState);
          return;
        }

        const nextState = await state.actions
          .webauthn_verify_attestation_response({
            public_key: attestationResponse,
          })
          .run();

        stateHandler[nextState.name](nextState);
      },
      async webauthn_credential_verification(state) {
        let attestationResponse: PublicKeyCredentialWithAttestationJSON;
        try {
          attestationResponse = await createWebauthnCredential({
            ...state.payload.creation_options,
            signal: _createAbortSignal(),
          });
        } catch (e) {
          const prevState = await state.actions.back(null).run();
          setLoadingAction(null);
          stateHandler[prevState.name](prevState);
          return;
        }

        const nextState = await state.actions
          .webauthn_verify_attestation_response({
            public_key: attestationResponse,
          })
          .run();

        stateHandler[nextState.name](nextState);
      },
      login_password(state) {
        setPage(<LoginPasswordPage state={state} />);
      },
      login_password_recovery(state) {
        setPage(<EditPasswordPage state={state} />);
      },
      login_method_chooser(state) {
        setPage(<LoginMethodChooserPage state={state} />);
      },
      registration_init(state) {
        setPage(<RegistrationInitPage state={state} />);
      },
      password_creation(state) {
        setPage(<CreatePasswordPage state={state} />);
      },
      success(state) {
        hanko.flow.client.processResponseHeadersOnLogin(
          "uuid", // TODO: replace, when the success-state payload includes the user details
          hanko.flow.client.response,
        );
        lastActionSucceeded();
        emitSuccessEvent("uuid");
      },
      profile_init(state) {
        setPage(
          <ProfilePage
            state={state}
            enablePasskeys={globalOptions.enablePasskeys}
          />,
        );
      },
      error(state) {
        setLoadingAction(null);
        setPage(<ErrorPage state={state} />);
      },
      onboarding_email(state) {
        setPage(<CreateEmailPage state={state} />);
      },
    }),
    [
      emitSuccessEvent,
      globalOptions.enablePasskeys,
      hanko,
      lastActionSucceeded,
      setLoadingAction,
    ],
  );

  const flowInit = useCallback(
    async (path: FlowPath) => {
      setLoadingAction("switch-flow");
      await hanko.flow.init(path, { ...stateHandler });
      setLoadingAction(null);
    },
    [stateHandler],
  );

  const init = useCallback(
    (compName: ComponentName) => {
      switch (compName) {
        case "auth":
          flowInit("/login").catch(handleError);
          break;
        case "login":
          flowInit("/login").catch(handleError);
          break;
        case "registration":
          flowInit("/registration").catch(handleError);
          break;
        case "profile":
          if (hanko.session.isValid()) {
            flowInit("/profile").catch(handleError);
          } else {
            setPage(<ErrorPage error={new UnauthorizedError()} />);
          }
          break;
      }
    },
    [flowInit],
  );

  useEffect(() => init(componentName), []);

  useEffect(() => {
    hanko.onAuthFlowCompleted((detail) => {
      dispatchEvent("onAuthFlowCompleted", detail);
    });

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
  }, []);

  const isWebAuthnSupported = WebauthnSupport.supported();

  return (
    <AppContext.Provider
      value={{
        init,
        initialComponentName: props.componentName,
        isDisabled,
        setUIState,
        setLoadingAction,
        setSucceededAction,
        uiState,
        hanko,
        lang: lang?.toString() || fallbackLanguage,
        prefilledEmail,
        prefilledUsername,
        componentName,
        setComponentName,
        experimentalFeatures,
        emitSuccessEvent,
        hidePasskeyButtonOnLogin,
        page,
        setPage,
        stateHandler,
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
