import { Fragment, h } from "preact";

import cx from "classnames";

import LoadingSpinner, {
  Props as LoadingSpinnerProps,
} from "../icons/LoadingSpinner";

import styles from "./styles.sass";
import { useCallback, useState } from "preact/compat";

type LoadingSpinnerPosition = "left" | "right";

export interface Props
  extends LoadingSpinnerProps,
    h.JSX.HTMLAttributes<HTMLButtonElement> {
  onClick(event: Event): void;
  dangerous?: boolean;
  loadingSpinnerPosition?: LoadingSpinnerPosition;
}

const Link = ({
  loadingSpinnerPosition,
  dangerous = false,
  onClick,
  ...props
}: Props) => {
  const [confirmationActive, setConfirmationActive] = useState<boolean>();

  const dangerousOnClick = (event: Event) => {
    event.preventDefault();
    setConfirmationActive(true);
  };

  const onCancel = (event: Event) => {
    event.preventDefault();
    setConfirmationActive(false);
  };

  const onConfirmation = useCallback(
    (event: Event) => {
      event.preventDefault();
      setConfirmationActive(false);
      onClick(event);
    },
    [onClick]
  );

  const renderLink = useCallback(
    () => (
      <Fragment>
        {confirmationActive ? (
          <Fragment>
            <Link onClick={onConfirmation}>✓ yes</Link>&nbsp;
            <Link onClick={onCancel}>✗ no</Link>
            &nbsp;
          </Fragment>
        ) : null}
        <button
          {...props}
          onClick={dangerous ? dangerousOnClick : onClick}
          disabled={confirmationActive || props.disabled || props.isLoading}
          // @ts-ignore
          part={"link"}
          className={cx(styles.link, dangerous ? styles.danger : null)}
        >
          {props.children}
        </button>
      </Fragment>
    ),
    [confirmationActive, dangerous, onClick, onConfirmation, props]
  );

  return (
    <Fragment>
      <span
        className={cx(
          styles.linkWrapper,
          loadingSpinnerPosition === "right" ? styles.reverse : null
        )}
        hidden={props.hidden}
      >
        {loadingSpinnerPosition && (props.isLoading || props.isSuccess) ? (
          <Fragment>
            <LoadingSpinner
              isLoading={props.isLoading}
              isSuccess={props.isSuccess}
              secondary={props.secondary}
              fadeOut
            />
            {renderLink()}
          </Fragment>
        ) : (
          <Fragment>{renderLink()}</Fragment>
        )}
      </span>
    </Fragment>
  );
};

export default Link;
