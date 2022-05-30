import * as preact from "preact";
import { ComponentChildren, createContext } from "preact";
import { useCallback, useContext, useEffect, useState } from "preact/compat";

import { HankoError, TooManyRequestsError } from "../../lib/Errors";

import { AppContext } from "./AppProvider";

interface Props {
  children: ComponentChildren;
}

interface Context {
  passwordInitialize: (userID: string) => Promise<HankoError>;
  passwordFinalize: (userID: string, password: string) => Promise<void>;
  passwordRetryAfter: number;
}

export const PasswordContext = createContext<Context>(null);

const PasswordProvider = ({ children }: Props) => {
  const { hanko } = useContext(AppContext);
  const [passwordRetryAfter, setPasswordRetryAfter] = useState<number>(0);

  const passwordInitialize = useCallback(
    (userID: string) => {
      return new Promise<HankoError>((resolve) => {
        const retryAfter = hanko.password.getRetryAfter(userID);

        setPasswordRetryAfter(retryAfter);
        resolve(retryAfter > 0 ? new TooManyRequestsError(retryAfter) : null);
      });
    },
    [hanko]
  );

  const passwordFinalize = useCallback(
    (userID: string, password: string) => {
      return new Promise<void>((resolve, reject) => {
        hanko.password
          .login(userID, password)
          .then(() => resolve())
          .catch((e) => {
            if (e instanceof TooManyRequestsError) {
              setPasswordRetryAfter(e.retryAfter);
            }

            return reject(e);
          });
      });
    },
    [hanko]
  );

  useEffect(() => {
    const timer =
      passwordRetryAfter > 0 &&
      setInterval(() => setPasswordRetryAfter(passwordRetryAfter - 1), 1000);

    return () => clearInterval(timer);
  }, [passwordRetryAfter]);

  return (
    <PasswordContext.Provider
      value={{ passwordInitialize, passwordFinalize, passwordRetryAfter }}
    >
      {children}
    </PasswordContext.Provider>
  );
};

export default PasswordProvider;
