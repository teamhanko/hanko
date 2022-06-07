import * as preact from "preact";
import { FunctionalComponent, RenderableProps } from "preact";
import { useContext } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";
import { RenderContext } from "../../contexts/PageProvider";

import Link, { Props as LinkProps } from "../Link";

const linkToEmailLogin = <P extends LinkProps>(
  LinkComponent: FunctionalComponent<LinkProps>
) => {
  return function LinkToEmailLogin(props: RenderableProps<P>) {
    const { t } = useContext(TranslateContext);
    const { renderLoginEmail } = useContext(RenderContext);

    const onClick = () => {
      renderLoginEmail();
    };

    return (
      <LinkComponent onClick={onClick} {...props}>
        {t("labels.back")}
      </LinkComponent>
    );
  };
};

export default linkToEmailLogin<LinkProps>(Link);
