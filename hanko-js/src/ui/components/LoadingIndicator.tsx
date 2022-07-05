import * as preact from "preact";
import { ComponentChildren } from "preact";

import Checkmark from "./Checkmark";
import LoadingWheel from "./LoadingWheel";

import styles from "./LoadingIndicator.sass";

export type Props = {
  children?: ComponentChildren;
  isLoading?: boolean;
  isSuccess?: boolean;
  fadeOut?: boolean;
  secondary?: boolean;
};

const LoadingIndicator = ({
  children,
  isLoading,
  isSuccess,
  fadeOut,
  secondary,
}: Props) => {
  return (
    <div className={styles.loadingIndicator}>
      {isLoading ? (
        <LoadingWheel />
      ) : isSuccess ? (
        <Checkmark fadeOut={fadeOut} secondary={secondary} />
      ) : (
        children
      )}
    </div>
  );
};

export default LoadingIndicator;
