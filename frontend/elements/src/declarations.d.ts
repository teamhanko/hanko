declare module "*.sass";

// eslint-disable-next-line no-unused-vars
interface Window {
  _hankoStyle: HTMLStyleElement;
}

declare module "react";

declare module "@denysvuika/preact-translate" {
  import { Context, h } from "preact";
  import { Dispatch, SetStateAction } from "preact/compat";

  interface TranslateParams {
    [key: string]: string | number;
  }

  interface LanguageData {
    [key: string]: any;
  }

  export const TranslateContext: Context<{
    lang: string;
    setLang: Dispatch<SetStateAction<string>>;
    t: (key: string, params?: TranslateParams) => string;
    isReady: boolean;
  }>;

  export interface TranslateProviderProps {
    root?: string;
    lang?: string;
    fallbackLang?: string;
    translations?: LanguageData;
    children?: any;
  }
  export const TranslateProvider: (
    props: TranslateProviderProps,
  ) => h.JSX.Element;
}
