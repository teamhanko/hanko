import { JSX, FunctionalComponent } from "preact";
import registerCustomElement from "@teamhanko/preact-custom-element";

import AppProvider from "./contexts/AppProvider";
import { Hanko } from "@teamhanko/hanko-frontend-sdk";

interface AdditionalProps {}

export interface HankoAuthAdditionalProps extends AdditionalProps {
  experimental?: string;
}

export interface HankoProfileAdditionalProps extends AdditionalProps {}

declare interface HankoAuthElementProps
  extends JSX.HTMLAttributes<HTMLElement>,
    HankoAuthAdditionalProps {}

declare interface HankoProfileElementProps
  extends JSX.HTMLAttributes<HTMLElement>,
    HankoProfileAdditionalProps {}

declare global {
  // eslint-disable-next-line no-unused-vars
  namespace JSX {
    // eslint-disable-next-line no-unused-vars
    interface IntrinsicElements {
      "hanko-auth": HankoAuthElementProps;
      "hanko-profile": HankoProfileElementProps;
    }
  }
}

export const HankoAuth = (props: HankoAuthElementProps) => (
  <AppProvider componentName={"auth"} {...props} hanko={hanko} />
);

export const HankoProfile = (props: HankoProfileElementProps) => (
  <AppProvider componentName={"profile"} {...props} hanko={hanko} />
);

export interface RegisterOptions {
  api: string;
  shadow?: boolean;
  injectStyles?: boolean;
}

interface InternalRegisterOptions extends RegisterOptions {
  tagName: string;
  entryComponent: FunctionalComponent<HankoAuthAdditionalProps>;
  observedAttributes: string[];
}

let hanko: Hanko;

interface ElementsRegisterReturn {
  hanko: Hanko;
}

const _register = async ({
  tagName,
  entryComponent,
  shadow = true,
  injectStyles = true,
  observedAttributes,
}: InternalRegisterOptions) => {
  if (!customElements.get(tagName)) {
    registerCustomElement(entryComponent, tagName, observedAttributes, {
      shadow,
    });
  }

  if (injectStyles) {
    await customElements.whenDefined(tagName);
    const elements = document.getElementsByTagName(tagName);
    const styles = window._hankoStyle;

    Array.from(elements).forEach((element) => {
      if (shadow) {
        const clonedStyles = styles.cloneNode(true);
        element.shadowRoot.appendChild(clonedStyles);
      } else {
        element.appendChild(styles);
      }
    });
  }
};

export const register = async (
  options: RegisterOptions
): Promise<ElementsRegisterReturn> => {
  createHankoClient(options.api);

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
  ]);

  return { hanko };
};

export const createHankoClient = (api: string) => {
  if (!hanko || hanko.api !== api) {
    hanko = new Hanko(api);
  }
  return hanko;
};
