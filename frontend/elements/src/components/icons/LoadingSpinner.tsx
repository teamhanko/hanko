import { ComponentChildren, Fragment } from "preact";
import styles from "./styles.sass";
import Icon from "./Icon";

export type Props = {
  children?: ComponentChildren;
  isLoading?: boolean;
  isSuccess?: boolean;
  fadeOut?: boolean;
  secondary?: boolean;
  hasIcon?: boolean;
};

const LoadingSpinner = ({
  children,
  isLoading,
  isSuccess,
  fadeOut,
  secondary,
  hasIcon,
}: Props) => {
  const partName = "loading-spinner-container";

  return (
    <Fragment>
      {isLoading ? (
        <div className={styles.loadingSpinnerWrapper} part={partName}>
          <Icon name={"spinner"} secondary={secondary} />
        </div>
      ) : isSuccess ? (
        <div className={styles.loadingSpinnerWrapper} part={partName}>
          <Icon name={"checkmark"} secondary={secondary} fadeOut={fadeOut} />
        </div>
      ) : (
        <div
          part={partName}
          className={
            hasIcon
              ? styles.loadingSpinnerWrapperIcon
              : styles.loadingSpinnerWrapper
          }
        >
          {children}
        </div>
      )}
    </Fragment>
  );
};

export default LoadingSpinner;
