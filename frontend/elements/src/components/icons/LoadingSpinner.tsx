import { ComponentChildren, Fragment } from "preact";
import styles from "./styles.sass";
import Icon from "./Icon";
import cx from "classnames";

export type Props = {
  children?: ComponentChildren;
  isLoading?: boolean;
  isSuccess?: boolean;
  fadeOut?: boolean;
  secondary?: boolean;
  hasIcon?: boolean;
  maxWidth?: boolean;
};

const LoadingSpinner = ({
  children,
  isLoading,
  isSuccess,
  fadeOut,
  secondary,
  hasIcon,
  maxWidth,
}: Props) => {
  return (
    <Fragment>
      {isLoading ? (
        <div className={cx(styles.loadingSpinnerWrapper, styles.centerContent, maxWidth && styles.maxWidth)}>
          <Icon name={"spinner"} secondary={secondary} />
        </div>
      ) : isSuccess ? (
        <div className={cx(styles.loadingSpinnerWrapper, styles.centerContent,  maxWidth && styles.maxWidth)}>
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
