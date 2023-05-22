import { JSX, FunctionalComponent } from "preact";
import registerCustomElement from "@teamhanko/preact-custom-element";

import AppProvider from "./contexts/AppProvider";
import { Hanko } from "@teamhanko/hanko-frontend-sdk";

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
  />
);

const HankoProfile = (props: HankoProfileElementProps) => (
  <AppProvider
    componentName={"profile"}
    {...props}
    hanko={hanko}
    injectStyles={injectStyles}
  />
);

const HankoEvents = (props: HankoProfileElementProps) => (
  <AppProvider componentName={"events"} {...props} hanko={hanko} />
);

export interface RegisterOptions {
  shadow?: boolean;
  injectStyles?: boolean;
}

interface InternalRegisterOptions extends RegisterOptions {
  tagName: string;
  entryComponent: FunctionalComponent<HankoAuthAdditionalProps>;
  observedAttributes: string[];
}

let hanko: Hanko;
let injectStyles: boolean;

export interface ElementsRegisterResult {
  hanko: Hanko;
}

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
  options: RegisterOptions = { shadow: true, injectStyles: true }
): Promise<ElementsRegisterResult> => {
  hanko = new Hanko(api);
  injectStyles = options.injectStyles;
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
