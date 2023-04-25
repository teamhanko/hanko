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

const HankoAuth = (props: HankoAuthElementProps) => (
  <AppProvider
    componentName={"auth"}
    {...props}
    hanko={hanko}
    injectStyles={injectStyles}
    translations={translations}
  />
);

const HankoProfile = (props: HankoProfileElementProps) => (
  <AppProvider
    componentName={"profile"}
    {...props}
    hanko={hanko}
    translations={translations}
    injectStyles={injectStyles}
  />
);

const HankoEvents = (props: HankoProfileElementProps) => (
  <AppProvider
    componentName={"events"}
    {...props}
    hanko={hanko}
    injectStyles={false}
    translations={null}
  />
);

export interface RegisterOptions {
  shadow?: boolean;
  injectStyles?: boolean;
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

let hanko: Hanko;
let injectStyles: boolean;
let translations: Translations;

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
    ...options,
  };
  hanko = new Hanko(api);
  injectStyles = options.injectStyles;
  translations = options.translations;
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

  return { hanko };
};

function deepMerge<T>(obj1: T, obj2: Partial<T>): T {
  const mergedObject = {} as T;

  for (const key in obj1) {
    if (Object.prototype.hasOwnProperty.call(obj1, key)) {
      const value1 = obj1[key];
      const value2 = obj2[key];

      if (
        typeof value1 === "object" &&
        value1 !== null &&
        typeof value2 === "object" &&
        value2 !== null
      ) {
        mergedObject[key] = deepMerge(value1, value2);
      } else {
        mergedObject[key] = value1;
      }
    }
  }

  for (const key in obj2) {
    if (Object.prototype.hasOwnProperty.call(obj2, key)) {
      const value1 = obj1[key];
      const value2 = obj2[key];

      if (typeof value1 === "undefined") {
        mergedObject[key] = value2;
      }
    }
  }

  return mergedObject;
}
