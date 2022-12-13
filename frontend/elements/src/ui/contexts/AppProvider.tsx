import * as preact from "preact";
import { ComponentChildren, createContext } from "preact";
import { useCallback, useMemo, useState } from "preact/compat";

import { Hanko, Config } from "@teamhanko/hanko-frontend-sdk";

type ExperimentalFeature = "conditionalMediation";
type ExperimentalFeatures = ExperimentalFeature[];

interface Props {
  api?: string;
  lang?: string;
  experimental?: string;
  children: ComponentChildren;
}

interface Context {
  config: Config;
  experimentalFeatures?: ExperimentalFeatures;
  configInitialize: () => Promise<Config>;
  hanko: Hanko;
}

export const AppContext = createContext<Context>(null);

const AppProvider = ({ api, children, experimental = "" }: Props) => {
  const [config, setConfig] = useState<Config>(null);

  const hanko = useMemo(() => {
    if (api.length) {
      return new Hanko(api, 13000);
    }
    return null;
  }, [api]);

  const experimentalFeatures = useMemo(
    () =>
      experimental
        .split(" ")
        .filter((feature) => feature.length)
        .map((feature) => feature as ExperimentalFeature),
    [experimental]
  );

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
    <AppContext.Provider
      value={{
        config,
        configInitialize,
        hanko,
        experimentalFeatures,
      }}
    >
      {children}
    </AppContext.Provider>
  );
};

export default AppProvider;
