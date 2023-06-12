import { JSX, FunctionalComponent } from "preact";
import registerCustomElement from "@teamhanko/preact-custom-element";
import AppProvider from "./contexts/AppProvider";
import { Hanko } from "@teamhanko/hanko-frontend-sdk";
import { defaultTranslations, Translations } from "./i18n/translations";

export interface HankoAuthAdditionalProps {
  experimental?: string;
}

export declare interface HankoAuthElementProps
  extends JSX.HTMLAttributes<HTMLElement>,
    HankoAuthAdditionalProps {}

export declare interface HankoProfileElementProps
  extends JSX.HTMLAttributes<HTMLElement> {}

export declare interface HankoEventsElementProps
  extends JSX.HTMLAttributes<HTMLElement> {}

declare global {
  // eslint-disable-next-line no-unused-vars
  namespace JSX {
    // eslint-disable-next-line no-unused-vars
    interface IntrinsicElements {
      "hanko-auth": HankoAuthElementProps;
      "hanko-profile": HankoProfileElementProps;
      "hanko-events": HankoEventsElementProps;
    }
  }
}

export interface RegisterOptions {
  shadow?: boolean;
  injectStyles?: boolean;
  enablePasskeys?: boolean;
  translations?: Translations;
  translationsLocation?: string;
  fallbackLanguage?: string;
}

export interface RegisterResult {
  hanko: Hanko;
}

interface InternalRegisterOptions extends RegisterOptions {
  tagName: string;
  entryComponent: FunctionalComponent<HankoAuthAdditionalProps>;
  observedAttributes: string[];
}

interface Global {
  hanko?: Hanko;
  injectStyles?: boolean;
  enablePasskeys?: boolean;
  translations?: Translations;
  translationsLocation?: string;
  fallbackLanguage?: string;
}

const global: Global = {};

const HankoAuth = (props: HankoAuthElementProps) => (
  <AppProvider
    componentName={"auth"}
    {...props}
    hanko={global.hanko}
    injectStyles={global.injectStyles}
    translations={global.translations}
    translationsLocation={global.translationsLocation}
    enablePasskeys={global.enablePasskeys}
    fallbackLanguage={global.fallbackLanguage}
  />
);

const HankoProfile = (props: HankoProfileElementProps) => (
  <AppProvider
    componentName={"profile"}
    {...props}
    hanko={global.hanko}
    injectStyles={global.injectStyles}
    translations={global.translations}
    translationsLocation={global.translationsLocation}
    enablePasskeys={global.enablePasskeys}
    fallbackLanguage={global.fallbackLanguage}
  />
);

const HankoEvents = (props: HankoProfileElementProps) => (
  <AppProvider componentName={"events"} {...props} hanko={global.hanko} />
);

const _register = async ({
  tagName,
  entryComponent,
  shadow = true,
  observedAttributes,
}: InternalRegisterOptions) => {
  if (!customElements.get(tagName)) {
    registerCustomElement(entryComponent, tagName, observedAttributes, {
      shadow,
    });
  }
};

export const register = async (
  api: string,
  options: RegisterOptions = {}
): Promise<RegisterResult> => {
  options = {
    shadow: true,
    injectStyles: true,
    enablePasskeys: true,
    translations: null,
    translationsLocation: "/i18n",
    fallbackLanguage: "en",
    ...options,
  };

  global.hanko = new Hanko(api);
  global.injectStyles = options.injectStyles;
  global.enablePasskeys = options.enablePasskeys;
  global.translations = options.translations || defaultTranslations;
  global.translationsLocation = options.translationsLocation;
  global.fallbackLanguage = options.fallbackLanguage;

  await Promise.all([
    _register({
      ...options,
      tagName: "hanko-auth",
      entryComponent: HankoAuth,
      observedAttributes: ["api", "lang", "experimental"],
    }),
    _register({
      ...options,
      tagName: "hanko-profile",
      entryComponent: HankoProfile,
      observedAttributes: ["api", "lang"],
    }),
    _register({
      ...options,
      tagName: "hanko-events",
      entryComponent: HankoEvents,
      observedAttributes: [],
    }),
  ]);

  return { hanko: global.hanko };
};
