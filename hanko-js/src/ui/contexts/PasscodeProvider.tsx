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
  passcodeTTL: number;
  passcodeResendAfter: number;
  passcodeInitialize: (userID: string) => Promise<HankoError>;
  passcodeResend: (userID: string) => Promise<void>;
  passcodeFinalize: (userID: string, passcode: string) => Promise<void>;
}

export const PasscodeContext = createContext<Context>(null);

const PasscodeProvider: FunctionalComponent = ({ children }: Props) => {
  const { hanko } = useContext(AppContext);

  const [passcodeTTL, setPasscodeTTL] = useState<number>(0);
  const [passcodeResendAfter, setPasscodeResendAfter] = useState<number>(0);
  const [passcodeIsActive, setPasscodeIsActive] = useState<boolean>(false);

  const passcodeResend = useCallback(
    (userID: string): Promise<void> => {
      return new Promise<void>((resolve, reject) => {
        hanko.passcode
          .initialize(userID)
          .then((passcode) => {
            setPasscodeTTL(passcode.ttl);

            return resolve();
          })
          .catch((e) => {
            if (e instanceof TooManyRequestsError) {
              setPasscodeResendAfter(e.retryAfter);
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
        const ttl = hanko.passcode.getTTL(userID);
        const resendAfter = hanko.passcode.getResendAfter(userID);

        setPasscodeTTL(ttl);
        setPasscodeResendAfter(resendAfter);

        if (ttl > 0) {
          setPasscodeIsActive(true);
          return resolve(null);
        } else if (resendAfter === 0) {
          passcodeResend(userID)
            .then(() => {
              setPasscodeIsActive(true);
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
          reject(new TooManyRequestsError(resendAfter));
        }
      });
    },
    [hanko.passcode, passcodeResend]
  );

  const passcodeFinalize = useCallback(
    (userID: string, code: string) => {
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
    },
    [hanko.passcode]
  );

  useEffect(() => {
    const timer =
      passcodeTTL > 0 &&
      setInterval(() => setPasscodeTTL(passcodeTTL - 1), 1000);
    return () => clearInterval(timer);
  }, [passcodeTTL]);

  useEffect(() => {
    const timer =
      passcodeResendAfter > 0 &&
      setInterval(() => setPasscodeResendAfter(passcodeResendAfter - 1), 1000);
    return () => clearInterval(timer);
  }, [passcodeResendAfter]);

  return (
    <PasscodeContext.Provider
      value={{
        passcodeTTL,
        passcodeIsActive,
        passcodeResendAfter,
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
