import * as preact from "preact";
import { ComponentChildren, createContext, FunctionalComponent } from "preact";
import { useCallback, useContext, useEffect, useState } from "preact/compat";
import {
  HankoError,
  TooManyRequestsError,
  MaxNumOfPasscodeAttemptsReachedError,
} from "../../lib/Errors";
import { AppContext } from "./AppProvider";

interface Props {
  children: ComponentChildren;
}

interface Context {
  passcodeIsActive: boolean;
  passcodeExpiry: number;
  passcodeRetryAfter: number;
  passcodeInitialize: (userID: string) => Promise<HankoError>;
  passcodeResend: (userID: string) => Promise<void>;
  passcodeFinalize: (userID: string, passcode: string) => Promise<void>;
}

export const PasscodeContext = createContext<Context>(null);

const PasscodeProvider: FunctionalComponent = ({ children }: Props) => {
  const { hanko } = useContext(AppContext);

  const [passcodeExpiry, setPasscodeExpiry] = useState<number>(0);
  const [passcodeRetryAfter, setPasscodeRetryAfter] = useState<number>(0);
  const [passcodeIsActive, setPasscodeIsActive] = useState<boolean>(false);

  const passcodeResend = useCallback(
    (userID: string): Promise<void> => {
      return new Promise<void>((resolve, reject) => {
        hanko.passcode
          .initialize(userID)
          .then((passcode) => {
            setPasscodeExpiry(passcode.ttl);
            setPasscodeIsActive(true);

            return resolve();
          })
          .catch((e) => {
            if (e instanceof TooManyRequestsError) {
              setPasscodeRetryAfter(e.retryAfter);
            }
            reject(e);
          });
      });
    },
    [hanko]
  );

  const passcodeInitialize = useCallback(
    (userID: string) => {
      return new Promise<HankoError>((resolve, reject) => {
        const expiry = hanko.passcode.getExpiry(userID);
        const retryAfter = hanko.passcode.getRetryAfter(userID);

        setPasscodeExpiry(expiry);
        setPasscodeRetryAfter(retryAfter);

        if (expiry > 0) {
          setPasscodeIsActive(true);
          return resolve(null);
        } else if (passcodeRetryAfter === 0) {
          passcodeResend(userID)
            .then(() => {
              return resolve(null);
            })
            .catch((e) => {
              if (e instanceof TooManyRequestsError) {
                resolve(e);
              } else {
                reject(e);
              }
            });
        } else {
          reject(new TooManyRequestsError(retryAfter));
        }
      });
    },
    [hanko.passcode, passcodeResend, passcodeRetryAfter]
  );

  const passcodeFinalize = (userID: string, code: string) => {
    return new Promise<void>((resolve, reject) => {
      hanko.passcode
        .finalize(userID, code)
        .then(() => resolve())
        .catch((e) => {
          if (e instanceof MaxNumOfPasscodeAttemptsReachedError) {
            setPasscodeIsActive(false);
          }
          reject(e);
        });
    });
  };

  useEffect(() => {
    const timer =
      passcodeExpiry > 0 &&
      setInterval(() => setPasscodeExpiry(passcodeExpiry - 1), 1000);
    return () => clearInterval(timer);
  }, [passcodeExpiry]);

  useEffect(() => {
    const timer =
      passcodeRetryAfter > 0 &&
      setInterval(() => setPasscodeRetryAfter(passcodeRetryAfter - 1), 1000);
    return () => clearInterval(timer);
  }, [passcodeRetryAfter]);

  return (
    <PasscodeContext.Provider
      value={{
        passcodeIsActive,
        passcodeExpiry,
        passcodeRetryAfter,
        passcodeResend,
        passcodeInitialize,
        passcodeFinalize,
      }}
    >
      {children}
    </PasscodeContext.Provider>
  );
};

export default PasscodeProvider;
