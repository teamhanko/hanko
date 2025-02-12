import { Fragment, h } from "preact";

import cx from "classnames";

import LoadingSpinner, {
  Props as LoadingSpinnerProps,
} from "../icons/LoadingSpinner";

import styles from "./styles.sass";
import { useCallback, useContext, useMemo, useState } from "preact/compat";
import { TranslateContext } from "@denysvuika/preact-translate";
import { AppContext } from "../../contexts/AppProvider";
import { useFlowEffects } from "../../contexts/UseFlowEffects";
import { Action } from "@teamhanko/hanko-frontend-sdk";

type LoadingSpinnerPosition = "left" | "right";

export interface Props
  extends LoadingSpinnerProps,
    h.JSX.HTMLAttributes<HTMLButtonElement> {
  onClick?(event: Event): void;
  dangerous?: boolean;
  loadingSpinnerPosition?: LoadingSpinnerPosition;
  flowAction?: Action<any>;
}

const Link = ({
  loadingSpinnerPosition,
  dangerous = false,
  onClick,
  flowAction,
  ...props
}: Props) => {
  const { t } = useContext(TranslateContext);
  const { uiState } = useContext(AppContext);
  const [confirmationActive, setConfirmationActive] = useState<boolean>();
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);

  onClick ||= async (e: Event) => {
    e.preventDefault();
    return await flowAction?.run();
  };

  useFlowEffects(flowAction, setIsLoading, setIsSuccess);

  let timeoutID: number;

  const dangerousOnClick = (event: Event) => {
    event.preventDefault();
    setConfirmationActive(true);
  };

  const onCancel = (event: Event) => {
    event.preventDefault();
    setConfirmationActive(false);
  };

  const loading = useMemo(
    () => isLoading || props.isLoading,
    [isLoading, props],
  );

  const success = useMemo(
    () => isSuccess || props.isSuccess,
    [isSuccess, props],
  );

  const hidden = useMemo(
    () => (flowAction && !flowAction.enabled) || props.hidden,
    [flowAction, props],
  );

  const onConfirmation = useCallback(
    (event: Event) => {
      event.preventDefault();
      setConfirmationActive(false);
      onClick(event);
    },
    [onClick],
  );

  const renderLink = useCallback(
    () =>
      !hidden ? (
        <Fragment>
          {confirmationActive ? (
            <Fragment>
              <Link onClick={onConfirmation}>{t("labels.yes")}</Link>
              &nbsp;/&nbsp;
              <Link onClick={onCancel}>{t("labels.no")}</Link>
              &nbsp;
            </Fragment>
          ) : null}
          <button
            {...props}
            onClick={dangerous ? dangerousOnClick : onClick}
            disabled={
              confirmationActive || props.disabled || uiState.isDisabled
            }
            part={"link"}
            className={cx(styles.link, dangerous ? styles.danger : null)}
          >
            {props.children}
          </button>
        </Fragment>
      ) : null,
    [
      hidden,
      uiState,
      confirmationActive,
      dangerous,
      onClick,
      onConfirmation,
      props,
      t,
    ],
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
          loadingSpinnerPosition === "right" ? styles.reverse : null,
        )}
        onMouseEnter={handleOnMouseEnter}
        onMouseLeave={handleOnMouseLeave}
      >
        {!confirmationActive && (loading || success) ? (
          <Fragment>
            <LoadingSpinner
              isLoading={loading}
              isSuccess={success}
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
