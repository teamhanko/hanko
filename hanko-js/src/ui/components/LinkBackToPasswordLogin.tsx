import * as preact from "preact";
import { useContext } from "preact/compat";

import { TranslateContext } from "@denysvuika/preact-translate";
import { RenderContext } from "../contexts/RenderProvider";

import Link from "./Link";

interface Props {
  userID: string;
  hidden?: boolean;
  disabled?: boolean;
}

const LinkBackToPasswordLogin = ({ userID, hidden, disabled }: Props) => {
  const { t } = useContext(TranslateContext);
  const { renderPassword } = useContext(RenderContext);

  const onClick = () => {
    renderPassword(userID);
  };

  return (
    <Link onClick={onClick} hidden={hidden} disabled={disabled}>
      {t("labels.back")}
    </Link>
  );
};

export default LinkBackToPasswordLogin;
