import * as preact from "preact";
import { ComponentChildren } from "preact";
import cx from "classnames";

import LoadingIndicator from "./LoadingIndicator";

import styles from "./Button.module.css";

type Props = {
  children: ComponentChildren;
  useSecondaryStyles?: boolean;
  isLoading?: boolean;
  isSuccess?: boolean;
  disabled?: boolean;
};

const Button = ({
  children,
  useSecondaryStyles,
  disabled,
  isLoading,
  isSuccess,
}: Props) => {
  return (
    <button
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
