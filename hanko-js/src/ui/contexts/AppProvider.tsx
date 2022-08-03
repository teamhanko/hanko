import * as preact from "preact";
import { ComponentChildren, createContext } from "preact";
import { useCallback, useMemo, useState } from "preact/compat";

import { Client, Config } from "../../lib/Client";

interface Props {
  api?: string;
  lang?: string;
  children: ComponentChildren;
}

interface Context {
  config: Config;
  configInitialize: () => Promise<Config>;
  hanko: Client;
}

export const AppContext = createContext<Context>(null);

const AppProvider = ({ api, children }: Props) => {
  const [config, setConfig] = useState<Config>(null);

  const hanko = useMemo(
    () => new Client(api.length ? api : "https://api.hanko.io", 13000),
    [api]
  );

  const configInitialize = useCallback(() => {
    return new Promise<Config>((resolve, reject) => {
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
