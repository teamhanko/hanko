import * as preact from "preact";
import { ComponentChildren } from "preact";

import cx from "classnames";

import Checkmark from "./Checkmark";

import styles from "./styles.sass";

export type Props = {
  children?: ComponentChildren;
  isLoading?: boolean;
  isSuccess?: boolean;
  fadeOut?: boolean;
  secondary?: boolean;
};

const LoadingSpinner = ({
  children,
  isLoading,
  isSuccess,
  fadeOut,
  secondary,
}: Props) => {
  return (
    <div className={styles.loadingSpinnerWrapper}>
      {isLoading ? (
        <div
          className={cx(styles.loadingSpinner, secondary && styles.secondary)}
        />
      ) : isSuccess ? (
        <Checkmark fadeOut={fadeOut} secondary={secondary} />
      ) : (
        children
      )}
    </div>
  );
};

export default LoadingSpinner;
