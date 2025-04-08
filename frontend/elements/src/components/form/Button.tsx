import { ComponentChildren } from "preact";
import {
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "preact/compat";

import cx from "classnames";

import styles from "./styles.sass";

import LoadingSpinner from "../icons/LoadingSpinner";
import Icon, { IconName } from "../icons/Icon";
import { AppContext } from "../../contexts/AppProvider";
import LastUsed from "./LastUsed";
import { useFlowEffects } from "../../hooks/UseFlowEffects";
import { useFormContext } from "./Form";

type Props = {
  title?: string;
  showSuccessIcon?: boolean;
  children: ComponentChildren;
  secondary?: boolean;
  dangerous?: boolean;
  isLoading?: boolean;
  isSuccess?: boolean;
  disabled?: boolean;
  autofocus?: boolean;
  showLastUsed?: boolean;
  onClick?: (event: Event) => void;
  icon?: IconName;
};

const Button = ({
  title,
  children,
  secondary,
  dangerous,
  autofocus,
  showLastUsed,
  onClick,
  icon,
  showSuccessIcon,
  ...props
}: Props) => {
  const ref = useRef(null);
  const { uiState } = useContext(AppContext);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);

  const { flowAction } = useFormContext();

  useFlowEffects(flowAction, setIsLoading, setIsSuccess);

  useEffect(() => {
    const { current: element } = ref;
    if (element && autofocus) {
      element.focus();
    }
  }, [autofocus]);

  const success = useMemo(
    () => showSuccessIcon && (isSuccess || props.isSuccess),
    [isSuccess, props, showSuccessIcon],
  );

  const disabled = useMemo(
    () => uiState.isDisabled || props.disabled,
    [props, uiState],
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
        isLoading={isLoading}
        isSuccess={success}
        secondary={true}
        hasIcon={!!icon}
        maxWidth
      >
        {icon ? (
          <Icon name={icon} secondary={secondary} disabled={disabled} />
        ) : null}
        <div className={styles.caption}>
          <span>{children}</span>
          {showLastUsed ? <LastUsed /> : null}
        </div>
      </LoadingSpinner>
    </button>
  );
};

export default Button;
