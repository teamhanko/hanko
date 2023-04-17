import { Fragment, h } from "preact";

import cx from "classnames";

import LoadingSpinner, {
  Props as LoadingSpinnerProps,
} from "../icons/LoadingSpinner";

import styles from "./styles.sass";
import { useCallback, useContext, useState } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";

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
  const { t } = useContext(TranslateContext);
  const [confirmationActive, setConfirmationActive] = useState<boolean>();
  let timeoutID: number;
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
            <Link onClick={onConfirmation}>{t("labels.yes")}</Link>&nbsp;/&nbsp;
            <Link onClick={onCancel}>{t("labels.no")}</Link>
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
    [confirmationActive, dangerous, onClick, onConfirmation, props, t]
  );

  const handleOnMouseEnter = () => {
    if (timeoutID) window.clearTimeout(timeoutID);
  };

  const handleOnMouseLeave = () => {
    timeoutID = window.setTimeout(() => {
      setConfirmationActive(false);
    }, 1000);
  };

  return (
    <Fragment>
      <span
        className={cx(
          styles.linkWrapper,
          loadingSpinnerPosition === "right" ? styles.reverse : null
        )}
        hidden={props.hidden}
        onMouseEnter={handleOnMouseEnter}
        onMouseLeave={handleOnMouseLeave}
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
