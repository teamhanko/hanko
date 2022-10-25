import * as preact from "preact";
import { FunctionalComponent, RenderableProps } from "preact";
import { useContext } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";
import { RenderContext } from "../../contexts/PageProvider";

import Link, { Props as LinkProps } from "../Link";

interface Props {
  userID: string;
}

const linkToPasswordLogin = <P extends Props & LinkProps>(
  LinkComponent: FunctionalComponent<LinkProps>
) => {
  return function LinkToPasswordLogin(props: RenderableProps<P>) {
    const { t } = useContext(TranslateContext);
    const { renderPassword, renderError } = useContext(RenderContext);

    const onClick = () => {
      renderPassword(props.userID).catch((e) => renderError(e));
    };

    return (
      <LinkComponent onClick={onClick} {...props}>
        {t("labels.back")}
      </LinkComponent>
    );
  };
};

export default linkToPasswordLogin<Props & LinkProps>(Link);
