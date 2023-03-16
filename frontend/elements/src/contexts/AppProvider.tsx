import { ComponentChildren, createContext, h } from "preact";
import { TranslateProvider } from "@denysvuika/preact-translate";

import {
  StateUpdater,
  useState,
  useCallback,
  useMemo,
  useRef,
  useEffect,
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

import { translations } from "../Translations";

import Container from "../components/wrapper/Container";

import InitPage from "../pages/InitPage";
import { JSXInternal } from "preact/src/jsx";
import SignalLike = JSXInternal.SignalLike;

type ExperimentalFeature = "conditionalMediation";
type ExperimentalFeatures = ExperimentalFeature[];
type ComponentName = "auth" | "profile";
type EventName = "hankoAuthSuccess" | "hankoUserDeleted";

interface Props {
  api?: string;
  lang?: string | SignalLike<string>;
  fallbackLang?: string;
  experimental?: string;
  componentName: ComponentName;
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

  emitEvent(eventName: EventName): void;
}

export const AppContext = createContext<Context>(null);

const AppProvider = ({
  api,
  lang,
  fallbackLang = "en",
  componentName,
  experimental = "",
}: Props) => {
  const ref = useRef<HTMLElement>(null);

  const initComponent = useMemo(() => <InitPage />, []);

  const hanko = useMemo(() => {
    if (api) {
      return new Hanko(api, 13000);
    }
    return null;
  }, [api]);

  const experimentalFeatures = useMemo(
    () =>
      experimental
        .split(" ")
        .filter((feature) => feature.length)
        .map((feature) => feature as ExperimentalFeature),
    [experimental]
  );

  const emitEvent = useCallback((eventName: EventName) => {
    const event = new Event(eventName, {
      bubbles: true,
      composed: true,
    });

    const fn = setTimeout(() => {
      ref.current.dispatchEvent(event);
    }, 500);

    return () => clearTimeout(fn);
  }, []);

  const [config, setConfig] = useState<Config>();
  const [userInfo, setUserInfo] = useState<UserInfo>(null);
  const [passcode, setPasscode] = useState<Passcode>();
  const [user, setUser] = useState<User>();
  const [emails, setEmails] = useState<Emails>();
  const [webauthnCredentials, setWebauthnCredentials] =
    useState<WebauthnCredentials>();
  const [page, setPage] = useState<h.JSX.Element>(initComponent);

  const listenEvent = useCallback((eventName: EventName, fn: () => void) => {
    document.addEventListener(eventName, fn);
    return () => document.removeEventListener(eventName, fn);
  }, []);

  const init = useCallback(() => {
    setPage(initComponent);
  }, [initComponent]);

  useEffect(() => {
    switch (componentName) {
      case "auth":
        return listenEvent("hankoUserDeleted", init);
      case "profile":
        return listenEvent("hankoAuthSuccess", init);
    }
  }, [componentName, init, listenEvent]);

  return (
    <AppContext.Provider
      value={{
        hanko,
        componentName,
        experimentalFeatures,
        emitEvent,
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
        fallbackLang={fallbackLang}
      >
        <Container ref={ref}>{page}</Container>
      </TranslateProvider>
    </AppContext.Provider>
  );
};

export default AppProvider;
