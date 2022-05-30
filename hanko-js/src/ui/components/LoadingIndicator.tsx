import * as preact from "preact";
import { ComponentChildren } from "preact";

import Checkmark from "./Checkmark";
import LoadingWheel from "./LoadingWheel";

import styles from "./LoadingIndicator.module.css";

export type Props = {
  children?: ComponentChildren;
  isLoading?: boolean;
  isSuccess?: boolean;
  fadeOut?: boolean;
  useSecondaryStyles?: boolean;
};

const LoadingIndicator = ({
  children,
  isLoading,
  isSuccess,
  fadeOut,
  useSecondaryStyles,
}: Props) => {
  return (
    <div className={styles.loadingIndicator}>
      {isLoading ? (
        <LoadingWheel />
      ) : isSuccess ? (
        <Checkmark fadeOut={fadeOut} useSecondaryStyles={useSecondaryStyles} />
      ) : (
        children
      )}
    </div>
  );
};

export default LoadingIndicator;
