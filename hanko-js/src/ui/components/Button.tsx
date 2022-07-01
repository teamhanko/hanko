import * as preact from "preact";
import { ComponentChildren } from "preact";
import { useEffect, useRef } from "preact/compat";

import cx from "classnames";

import LoadingIndicator from "./LoadingIndicator";
import styles from "./Button.sass";

type Props = {
  children: ComponentChildren;
  secondary?: boolean;
  isLoading?: boolean;
  isSuccess?: boolean;
  disabled?: boolean;
  autofocus?: boolean;
};

const Button = ({
  children,
  secondary,
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
      // @ts-ignore
      part={secondary ? "button secondary-button" : "button primary-button"}
      ref={ref}
      type={"submit"}
      disabled={disabled || isLoading || isSuccess}
      className={cx(
        styles.button,
        secondary ? styles.secondary : styles.primary
      )}
    >
      <LoadingIndicator
        isLoading={isLoading}
        isSuccess={isSuccess}
        secondary={!secondary}
      >
        {children}
      </LoadingIndicator>
    </button>
  );
};

export default Button;
