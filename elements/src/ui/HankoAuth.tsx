import * as preact from "preact";
import registerCustomElement from "preact-custom-element";
import { Fragment } from "preact";

import { TranslateProvider } from "@denysvuika/preact-translate";

import PageProvider from "./contexts/PageProvider";
import AppProvider from "./contexts/AppProvider";
import UserProvider from "./contexts/UserProvider";
import PasscodeProvider from "./contexts/PasscodeProvider";
import PasswordProvider from "./contexts/PasswordProvider";

import { translations } from "./Translations";

interface Props {
  api: string;
  lang?: string;
}

declare interface HankoAuthElement
  extends preact.JSX.HTMLAttributes<HTMLElement>,
    Props {}

declare global {
  // eslint-disable-next-line no-unused-vars
  namespace JSX {
    // eslint-disable-next-line no-unused-vars
    interface IntrinsicElements {
      "hanko-auth": HankoAuthElement;
    }
  }
}

export const HankoAuth = ({ api = "", lang = "en" }: Props) => {
  return (
    <Fragment>
      <AppProvider api={api}>
        <TranslateProvider translations={translations} fallbackLang={"en"}>
          <UserProvider>
            <PasswordProvider>
              <PasscodeProvider>
                <PageProvider lang={lang} />
              </PasscodeProvider>
            </PasswordProvider>
          </UserProvider>
        </TranslateProvider>
      </AppProvider>
    </Fragment>
  );
};

export interface RegisterOptions {
  shadow?: boolean;
  injectStyles?: boolean;
}

export const register = ({
  shadow = true,
  injectStyles = true,
}: RegisterOptions): Promise<void> => {
  const tagName = "hanko-auth";

  return new Promise<void>((resolve, reject) => {
    if (!customElements.get(tagName)) {
      registerCustomElement(HankoAuth, tagName, ["api", "lang"], {
        shadow,
      });
    }

    if (injectStyles) {
      customElements
        .whenDefined(tagName)
        .then((_) => {
          const elements = document.getElementsByTagName(tagName);

          Array.from(elements).forEach((element) => {
            if (shadow) {
              element.shadowRoot.appendChild(window._hankoStyle);
            } else {
              element.appendChild(window._hankoStyle);
            }
          });

          return resolve();
        })
        .catch((e) => {
          reject(e);
        });
    } else {
      return resolve();
    }
  });
};
