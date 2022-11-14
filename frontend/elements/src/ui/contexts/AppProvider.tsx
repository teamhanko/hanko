import * as preact from "preact";
import { ComponentChildren, createContext } from "preact";
import { useCallback, useMemo, useState } from "preact/compat";

import { Hanko, Config } from "@teamhanko/hanko-frontend-sdk";

interface Props {
  api?: string;
  lang?: string;
  children: ComponentChildren;
}

interface Context {
  config: Config;
  configInitialize: () => Promise<Config>;
  hanko: Hanko;
}

export const AppContext = createContext<Context>(null);

const AppProvider = ({ api, children }: Props) => {
  const [config, setConfig] = useState<Config>(null);

  const hanko = useMemo(() => {
    if (api.length) {
      return new Hanko(api, 13000);
    }
    return null;
  }, [api]);

  const configInitialize = useCallback(() => {
    return new Promise<Config>((resolve, reject) => {
      if (!hanko) {
        return;
      }

      hanko.config
        .get()
        .then((c) => {
          setConfig(c);

          return resolve(c);
        })
        .catch((e) => reject(e));
    });
  }, [hanko]);

  return (
    <AppContext.Provider value={{ config, configInitialize, hanko }}>
      {children}
    </AppContext.Provider>
  );
};

export default AppProvider;
