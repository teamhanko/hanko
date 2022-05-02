import * as preact from "preact";
import { ComponentChildren } from "preact";

import Checkmark from "./Checkmark";
import LoadingWheel from "./LoadingWheel";

import styles from "./LoadingIndicator.module.css";

type Props = {
  children?: ComponentChildren;
  isLoading?: boolean;
  isSuccess?: boolean;
  fadeOutCheckmark?: boolean;
  useSecondaryStyles?: boolean;
};

const LoadingIndicator = ({
  children,
  isLoading,
  isSuccess,
  fadeOutCheckmark,
  useSecondaryStyles,
}: Props) => {
  return (
    <div className={styles.loadingIndicator}>
      {isLoading ? (
        <LoadingWheel />
      ) : isSuccess ? (
        <Checkmark
          fadeOut={fadeOutCheckmark}
          useSecondaryStyles={useSecondaryStyles}
        />
      ) : (
        children
      )}
    </div>
  );
};

export default LoadingIndicator;
