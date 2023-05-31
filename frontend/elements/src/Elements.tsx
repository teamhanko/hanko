import { JSX, FunctionalComponent } from "preact";
import registerCustomElement from "@teamhanko/preact-custom-element";
import AppProvider from "./contexts/AppProvider";
import { Hanko } from "@teamhanko/hanko-frontend-sdk";
import { defaultTranslations, Translations } from "./Translations";

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
  translations?: Partial<Translations>;
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
  translations?: Partial<Translations>;
}

const global: Global = {};

const HankoAuth = (props: HankoAuthElementProps) => (
  <AppProvider
    componentName={"auth"}
    {...props}
    hanko={global.hanko}
    injectStyles={global.injectStyles}
    enablePasskeys={global.enablePasskeys}
    translations={global.translations}
  />
);

const HankoProfile = (props: HankoProfileElementProps) => (
  <AppProvider
    componentName={"profile"}
    {...props}
    hanko={global.hanko}
    injectStyles={global.injectStyles}
    enablePasskeys={global.enablePasskeys}
    translations={global.translations}
  />
);

const HankoEvents = (props: HankoProfileElementProps) => (
  <AppProvider
    componentName={"events"}
    {...props}
    hanko={global.hanko}
    translations={null}
  />
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
    translations: { ...defaultTranslations },
    enablePasskeys: true,
    ...options,
  };

  global.hanko = new Hanko(api);
  global.injectStyles = options.injectStyles;
  global.enablePasskeys = options.enablePasskeys;
  global.translations = options.translations;

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
