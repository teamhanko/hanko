import * as preact from "preact";
import { FunctionalComponent, RenderableProps } from "preact";
import cx from "classnames";

import Link, { Props as LinkProps } from "../Link";
import LoadingIndicator, {
  Props as LoadingIndicatorProps,
} from "../LoadingIndicator";

import styles from "./withLoadingIndicator.sass";

export interface Props {
  swap?: boolean;
}

const linkWithLoadingIndicator = <
  P extends Props & LinkProps & LoadingIndicatorProps
>(
  LinkComponent: FunctionalComponent
) => {
  return function LinkWithLoadingIndicator(props: RenderableProps<P>) {
    return (
      <span
        className={cx(
          styles.linkWithLoadingIndicator,
          props.swap ? styles.swap : null
        )}
        hidden={props.hidden}
      >
        <LoadingIndicator
          isLoading={props.isLoading}
          isSuccess={props.isSuccess}
          fadeOut
        />
        <LinkComponent {...props} />
      </span>
    );
  };
};

export default linkWithLoadingIndicator<
  Props & LinkProps & LoadingIndicatorProps
>(Link);
