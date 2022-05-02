import * as preact from "preact";
import { ComponentChildren } from "preact";
import cx from "classnames";

import Link from "./Link";
import LoadingIndicator from "./LoadingIndicator";

import styles from "./LinkWithLoadingIndicator.module.css";

type Props = {
  children?: ComponentChildren;
  onClick: (event: Event) => void;
  disabled?: boolean;
  isLoading?: boolean;
  isSuccess?: boolean;
  reverse?: boolean;
  hidden?: boolean;
};

const LinkWithLoadingIndicator = ({
  children,
  onClick,
  disabled,
  isLoading,
  isSuccess,
  reverse,
  hidden,
}: Props) => {
  return (
    <span
      className={cx(
        styles.linkWithLoadingIndicator,
        reverse ? styles.reverse : null
      )}
      hidden={hidden}
    >
      <LoadingIndicator
        isLoading={isLoading}
        isSuccess={isSuccess}
        useSecondaryStyles
        fadeOutCheckmark
      />
      <Link onClick={onClick} disabled={disabled} hidden={hidden}>
        {children}
      </Link>
    </span>
  );
};

export default LinkWithLoadingIndicator;
