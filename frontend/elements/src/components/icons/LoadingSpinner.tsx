import * as preact from "preact";
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
  return (
    <Fragment>
      {isLoading ? (
        <div className={styles.loadingSpinnerWrapper}>
          <Icon name={"spinner"} secondary={secondary} />
        </div>
      ) : isSuccess ? (
        <div className={styles.loadingSpinnerWrapper}>
          <Icon name={"checkmark"} secondary={secondary} fadeOut={fadeOut} />
        </div>
      ) : (
        <div
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
