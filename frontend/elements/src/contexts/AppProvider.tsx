import { ComponentChildren, createContext, h } from "preact";
import { TranslateProvider } from "@denysvuika/preact-translate";

import {
  StateUpdater,
  useState,
  useCallback,
  useMemo,
  useRef,
  useEffect,
  Fragment,
} from "preact/compat";

import {
  Hanko,
  User,
  UserInfo,
  Passcode,
  Emails,
  Config,
  WebauthnCredentials,
} from "@teamhanko/hanko-frontend-sdk";

import { Translations } from "../i18n/translations";

import Container from "../components/wrapper/Container";

import InitPage from "../pages/InitPage";
import { JSXInternal } from "preact/src/jsx";
import SignalLike = JSXInternal.SignalLike;

type ExperimentalFeature = "conditionalMediation";
type ExperimentalFeatures = ExperimentalFeature[];
export type ComponentName = "auth" | "profile" | "events";

export interface GlobalOptions {
  hanko?: Hanko;
  injectStyles?: boolean;
  enablePasskeys?: boolean;
  hidePasskeyButtonOnLogin?: boolean;
  translations?: Translations;
  translationsLocation?: string;
  fallbackLanguage?: string;
}

interface Props {
  lang?: string | SignalLike<string>;
  experimental?: string;
  componentName: ComponentName;
  globalOptions: GlobalOptions;
  children?: ComponentChildren;
}

interface States {
  config: Config;
  setConfig: StateUpdater<Config>;
  userInfo: UserInfo;
  setUserInfo: StateUpdater<UserInfo>;
  passcode: Passcode;
  setPasscode: StateUpdater<Passcode>;
  user: User;
  setUser: StateUpdater<User>;
  emails: Emails;
  setEmails: StateUpdater<Emails>;
  webauthnCredentials: WebauthnCredentials;
  setWebauthnCredentials: StateUpdater<WebauthnCredentials>;
  page: h.JSX.Element;
  setPage: StateUpdater<h.JSX.Element>;
}

interface Context extends States {
  hanko: Hanko;
  componentName: ComponentName;
  experimentalFeatures?: ExperimentalFeatures;
  emitSuccessEvent: (userID: string) => void;
  enablePasskeys: boolean;
  lang: string;
  hidePasskeyButtonOnLogin: boolean;
}

export const AppContext = createContext<Context>(null);
const AppProvider = ({
  lang,
  componentName,
  experimental = "",
  globalOptions,
}: Props) => {
  const {
    hanko,
    injectStyles,
    enablePasskeys,
    hidePasskeyButtonOnLogin,
    translations,
    translationsLocation,
    fallbackLanguage,
  } = globalOptions;
  const ref = useRef<HTMLElement>(null);
  const experimentalFeatures = useMemo(
    () =>
      experimental
        .split(" ")
        .filter((feature) => feature.length)
        .map((feature) => feature as ExperimentalFeature),
    [experimental]
  );

  const initComponent = useMemo(() => <InitPage />, []);
  const [config, setConfig] = useState<Config>();
  const [userInfo, setUserInfo] = useState<UserInfo>(null);
  const [passcode, setPasscode] = useState<Passcode>();
  const [user, setUser] = useState<User>();
  const [emails, setEmails] = useState<Emails>();
  const [webauthnCredentials, setWebauthnCredentials] =
    useState<WebauthnCredentials>();
  const [page, setPage] = useState<h.JSX.Element>(initComponent);

  const init = useCallback(() => {
    setPage(initComponent);
  }, [initComponent]);

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
    [hanko]
  );

  useMemo(() => {
    switch (componentName) {
      case "auth":
        hanko.onUserLoggedOut(init);
        hanko.onSessionExpired(init);
        hanko.onUserDeleted(init);
        break;
      case "profile":
        hanko.onSessionCreated(init);
        break;
    }
  }, [componentName, hanko, init]);

  const dispatchEvent = function <T>(type: string, detail?: T) {
    ref.current?.dispatchEvent(
      new CustomEvent<T>(type, {
        detail,
        bubbles: false,
        composed: true,
      })
    );
  };

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

  return (
    <AppContext.Provider
      value={{
        hanko,
        lang: lang?.toString(),
        componentName,
        experimentalFeatures,
        emitSuccessEvent,
        enablePasskeys,
        hidePasskeyButtonOnLogin,
        config,
        setConfig,
        userInfo,
        setUserInfo,
        passcode,
        setPasscode,
        user,
        setUser,
        emails,
        setEmails,
        webauthnCredentials,
        setWebauthnCredentials,
        page,
        setPage,
      }}
    >
      <TranslateProvider
        translations={translations}
        lang={lang?.toString()}
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
