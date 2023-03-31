import { ComponentChildren, createContext, h } from "preact";
import { TranslateProvider } from "@denysvuika/preact-translate";

import {
  StateUpdater,
  useState,
  useCallback,
  useMemo,
  useRef,
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
  emitSuccessEvent: (userID: string) => void;
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
        hanko.onSessionRemoved(init);
        hanko.onUserDeleted(init);
        break;
      case "profile":
        hanko.onSessionCreated(init);
        hanko.onSessionRemoved(init);
        break;
    }
  }, [componentName, hanko, init]);

  return (
    <AppContext.Provider
      value={{
        hanko,
        componentName,
        experimentalFeatures,
        emitSuccessEvent,
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
