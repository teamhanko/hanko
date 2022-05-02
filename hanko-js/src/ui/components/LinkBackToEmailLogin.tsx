import * as preact from "preact";
import { useContext, useState } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";
import { RenderContext } from "../contexts/RenderProvider";

import LinkWithLoadingIndicator from "./LinkWithLoadingIndicator";

interface Props {
  disabled?: boolean;
  hidden?: boolean;
  reverse?: boolean;
}

const LinkBackToEmailLogin = ({ disabled, hidden, reverse = true }: Props) => {
  const { t } = useContext(TranslateContext);
  const { renderLoginEmail } = useContext(RenderContext);

  const [isLoading, setIsLoading] = useState<boolean>(false);

  const onClick = () => {
    setIsLoading(true);
    renderLoginEmail();
  };

  return (
    <LinkWithLoadingIndicator
      isLoading={isLoading}
      onClick={onClick}
      disabled={disabled}
      hidden={hidden}
      reverse={reverse}
    >
      {t("labels.back")}
    </LinkWithLoadingIndicator>
  );
};

export default LinkBackToEmailLogin;
