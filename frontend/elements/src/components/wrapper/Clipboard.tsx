import { ComponentChildren, h } from "preact";

import styles from "./styles.sass";
import { useContext, useState } from "preact/compat";
import Icon from "../icons/Icon";
import { TranslateContext } from "@denysvuika/preact-translate";

type Props = {
  text: string;
  children: ComponentChildren;
};

const Clipboard = ({ children, text }: Props) => {
  const { t } = useContext(TranslateContext);
  const [isCopied, setIsCopied] = useState(false);

  const copyToClipboard = async (event: Event) => {
    event.preventDefault();
    try {
      await navigator.clipboard.writeText(text);
      setIsCopied(true);
      setTimeout(() => setIsCopied(false), 1500); // Reset after 1.5 seconds
    } catch (err) {
      console.error("Failed to copy: ", err);
    }
  };

  return (
    <section className={styles.clipboardContainer}>
      <div>{children}&nbsp;</div>
      <div className={styles.clipboardIcon} onClick={copyToClipboard}>
        {isCopied ? (
          <span>- {t("labels.copied")}</span>
        ) : (
          <Icon name={"copy"} secondary size={13} />
        )}
      </div>
    </section>
  );
};

export default Clipboard;
