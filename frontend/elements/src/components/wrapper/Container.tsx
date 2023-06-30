import { ComponentChildren, h } from "preact";
import { forwardRef, useContext, useEffect } from "preact/compat";

import styles from "./styles.sass";
import { AppContext } from "../../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

interface Props extends h.JSX.HTMLAttributes<HTMLElement> {
  children: ComponentChildren;
}

const Container = forwardRef<HTMLElement>((props: Props, ref) => {
  const { lang } = useContext(AppContext);
  const { setLang } = useContext(TranslateContext);

  useEffect(() => {
    setLang(lang);
  }, [lang, setLang]);

  return (
    <section
      // @ts-ignore
      part={"container"}
      className={styles.container}
      ref={ref}
    >
      {props.children}
    </section>
  );
});

export default Container;
