import { ComponentChildren } from "preact";
import { useContext, useEffect, useMemo, useRef } from "preact/compat";

import cx from "classnames";

import styles from "./styles.sass";

import LoadingSpinner from "../icons/LoadingSpinner";
import Icon, { IconName } from "../icons/Icon";
import { AppContext, UIAction } from "../../contexts/AppProvider";

type Props = {
  uiAction?: UIAction;
  title?: string;
  children: ComponentChildren;
  secondary?: boolean;
  dangerous?: boolean;
  isLoading?: boolean;
  isSuccess?: boolean;
  disabled?: boolean;
  autofocus?: boolean;
  onClick?: (event: Event) => void;
  icon?: IconName;
};

const Button = ({
  uiAction,
  title,
  children,
  secondary,
  dangerous,
  autofocus,
  onClick,
  icon,
  ...props
}: Props) => {
  const ref = useRef(null);
  const { uiState, isDisabled } = useContext(AppContext);

  useEffect(() => {
    const { current: element } = ref;
    if (element && autofocus) {
      element.focus();
    }
  }, [autofocus]);

  const loading = useMemo(
    () => (uiAction && uiState.loadingAction === uiAction) || props.isLoading,
    [props, uiAction, uiState],
  );

  const success = useMemo(
    () => (uiAction && uiState.succeededAction === uiAction) || props.isSuccess,
    [props, uiAction, uiState],
  );

  const disabled = useMemo(
    () => isDisabled || props.disabled,
    [props, isDisabled],
  );

  return (
    <button
      // @ts-ignore
      part={
        dangerous
          ? "button dangerous-button"
          : secondary
          ? "button secondary-button"
          : "button primary-button"
      }
      title={title}
      ref={ref}
      type={"submit"}
      disabled={disabled}
      onClick={onClick}
      className={cx(
        styles.button,
        dangerous
          ? styles.dangerous
          : secondary
          ? styles.secondary
          : styles.primary,
      )}
    >
      <LoadingSpinner
        isLoading={loading}
        isSuccess={success}
        secondary={true}
        hasIcon={!!icon}
        maxWidth
      >
        {icon ? (
          <Icon name={icon} secondary={secondary} disabled={disabled} />
        ) : null}
        {children}
      </LoadingSpinner>
    </button>
  );
};

export default Button;
