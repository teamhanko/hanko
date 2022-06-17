import * as preact from "preact";
import register from "preact-custom-element";
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

const HankoAuth = ({ api, lang }: Props) => {
  return (
    <Fragment>
      <style
        dangerouslySetInnerHTML={{ __html: window._hankoStyle.innerHTML }}
      />
      <AppProvider api={api}>
        <TranslateProvider
          translations={translations}
          lang={lang}
          fallbackLang={"en"}
        >
          <UserProvider>
            <PasswordProvider>
              <PasscodeProvider>
                <PageProvider />
              </PasscodeProvider>
            </PasswordProvider>
          </UserProvider>
        </TranslateProvider>
      </AppProvider>
    </Fragment>
  );
};

register(HankoAuth, "hanko-auth", ["api", "lang"], {
  shadow: true,
});

export default HankoAuth;
