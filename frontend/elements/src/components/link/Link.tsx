import * as preact from "preact";
import { Fragment, h } from "preact";

import cx from "classnames";

import LoadingSpinner, {
  Props as LoadingSpinnerProps,
} from "../icons/LoadingSpinner";

import styles from "./styles.sass";

type LoadingSpinnerPosition = "left" | "right";

export interface Props
  extends LoadingSpinnerProps,
    h.JSX.HTMLAttributes<HTMLButtonElement> {
  loadingSpinnerPosition?: LoadingSpinnerPosition;
}

const Link = ({ loadingSpinnerPosition, ...props }: Props) => {
  const renderLink = () => (
    <button
      {...props}
      // @ts-ignore
      part={"link"}
      className={cx(styles.link, props.disabled ? styles.disabled : null)}
    >
      {props.children}
    </button>
  );

  return (
    <Fragment>
      {loadingSpinnerPosition ? (
        <span
          className={cx(
            styles.linkWrapper,
            loadingSpinnerPosition === "right" ? styles.reverse : null
          )}
          hidden={props.hidden}
        >
          <LoadingSpinner
            isLoading={props.isLoading}
            isSuccess={props.isSuccess}
            secondary={props.secondary}
            fadeOut
          />
          {renderLink()}
        </span>
      ) : (
        <Fragment>{renderLink()}</Fragment>
      )}
    </Fragment>
  );
};

export default Link;
