import * as preact from "preact";
import { ComponentChildren } from "preact";
import { useEffect, useRef } from "preact/compat";

import cx from "classnames";

import LoadingIndicator from "./LoadingIndicator";
import styles from "./Button.sass";

type Props = {
  children: ComponentChildren;
  useSecondaryStyles?: boolean;
  isLoading?: boolean;
  isSuccess?: boolean;
  disabled?: boolean;
  autofocus?: boolean;
};

const Button = ({
  children,
  useSecondaryStyles,
  disabled,
  isLoading,
  isSuccess,
  autofocus,
}: Props) => {
  const ref = useRef(null);

  useEffect(() => {
    const { current: element } = ref;
    if (element && autofocus) {
      element.focus();
    }
  }, [autofocus]);

  return (
    <button
      ref={ref}
      type={"submit"}
      disabled={disabled || isLoading || isSuccess}
      className={cx(
        styles.button,
        useSecondaryStyles ? styles.secondary : styles.primary
      )}
    >
      <LoadingIndicator
        isLoading={isLoading}
        isSuccess={isSuccess}
        useSecondaryStyles={useSecondaryStyles}
      >
        {children}
      </LoadingIndicator>
    </button>
  );
};

export default Button;
