import { ComponentChildren, toChildArray } from "preact";

import styles from "./styles.sass";
import cx from "classnames";
import { Action } from "@teamhanko/hanko-frontend-sdk";
import { useContext, createContext } from "preact/compat";

type Props = {
  onSubmit?: (event: Event) => void;
  children: ComponentChildren;
  hidden?: boolean;
  maxWidth?: boolean;
  flowAction?: Action<any>;
};

type FormContextType = {
  flowAction?: Action<any>;
};

export const FormContext = createContext<FormContextType>({});

export const useFormContext = () => useContext(FormContext);

const Form = ({
  onSubmit,
  children,
  hidden = false,
  maxWidth,
  flowAction,
}: Props) => {
  const defaultOnSubmit = async (event: Event) => {
    event.preventDefault();
    return await flowAction.run();
  };

  // Cast Provider to any to bypass strict JSX return type check (TS2786)
  // TODO: Find out why, we this need to be casted to any for the build to work.
  const FormContextProviderAny = FormContext.Provider as any;

  return (
    <FormContextProviderAny value={{ flowAction }}>
      {flowAction && flowAction.enabled && !hidden ? (
        <form onSubmit={onSubmit || defaultOnSubmit} className={styles.form}>
          <ul className={styles.ul}>
            {toChildArray(children).map((child, index) => (
              <li
                part={"form-item"}
                className={cx(styles.li, maxWidth ? styles.maxWidth : null)}
                key={index}
              >
                {child}
              </li>
            ))}
          </ul>
        </form>
      ) : null}
    </FormContextProviderAny>
  );
};

export default Form;
