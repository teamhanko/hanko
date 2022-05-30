import * as preact from "preact";
import { ComponentChildren } from "preact";
import {
  createContext,
  StateUpdater,
  useCallback,
  useContext,
  useState,
} from "preact/compat";

import { User } from "../../lib/HankoClient";

import { AppContext } from "./AppProvider";

interface Props {
  children: ComponentChildren;
}

interface Context {
  user: User;
  email: string;
  setEmail: StateUpdater<string>;
  userInitialize: () => Promise<User>;
}

export const UserContext = createContext<Context>(null);

const UserProvider = ({ children }: Props) => {
  const { hanko } = useContext(AppContext);
  const [user, setUser] = useState<User>(null);
  const [email, setEmail] = useState<string>(null);

  const userInitialize = useCallback(() => {
    return new Promise<User>((resolve, reject) => {
      hanko.user
        .getCurrent()
        .then((u) => {
          setUser(u);

          return resolve(u);
        })
        .catch((e) => {
          reject(e);
        });
    });
  }, [hanko]);

  return (
    <UserContext.Provider
      value={{
        user,
        email,
        setEmail,
        userInitialize,
      }}
    >
      {children}
    </UserContext.Provider>
  );
};

export default UserProvider;
