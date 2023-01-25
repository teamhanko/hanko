import * as preact from "preact";
import { ComponentChildren } from "preact";
import { useEffect, useRef } from "preact/compat";

import cx from "classnames";

import styles from "./styles.sass";

import LoadingSpinner from "../icons/LoadingSpinner";

type Props = {
  title?: string
  children: ComponentChildren;
  secondary?: boolean;
  isLoading?: boolean;
  isSuccess?: boolean;
  disabled?: boolean;
  autofocus?: boolean;
};

const Button = ({
  title,
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
      title={title}
      ref={ref}
      type={"submit"}
      disabled={disabled || isLoading || isSuccess}
      className={cx(
        styles.button,
        secondary ? styles.secondary : styles.primary
      )}
    >
      <LoadingSpinner
        isLoading={isLoading}
        isSuccess={isSuccess}
        secondary={true}
      >
        {children}
      </LoadingSpinner>
    </button>
  );
};

export default Button;
