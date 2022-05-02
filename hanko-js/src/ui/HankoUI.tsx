import * as preact from "preact";
import register from "preact-custom-element";
import { Fragment } from "preact";

import { TranslateProvider } from "@denysvuika/preact-translate";
import RenderProvider from "./contexts/RenderProvider";
import AppProvider from "./contexts/AppProvider";
import UserProvider from "./contexts/UserProvider";
import PasscodeProvider from "./contexts/PasscodeProvider";
import PasswordProvider from "./contexts/PasswordProvider";

import { translations } from "./Translations";

interface Props {
  api: string;
  lang?: string;
}

const HankoUI = ({ api, lang }: Props) => {
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
                <RenderProvider />
              </PasscodeProvider>
            </PasswordProvider>
          </UserProvider>
        </TranslateProvider>
      </AppProvider>
    </Fragment>
  );
};

register(HankoUI, "x-hanko", ["api", "lang"], {
  shadow: true,
});

export default HankoUI;
