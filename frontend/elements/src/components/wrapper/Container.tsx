import { ComponentChildren, h, HTMLAttributes, JSX } from "preact";
import { forwardRef, useContext, useEffect } from "preact/compat";

import styles from "./styles.sass";
import { AppContext } from "../../contexts/AppProvider";
import { TranslateContext } from "@denysvuika/preact-translate";

interface Props extends HTMLAttributes {
  children: ComponentChildren;
}

const Container = forwardRef<HTMLElement>((props: Props, ref)=> {
  const { lang, hanko, setHanko } = useContext(AppContext);
  const { setLang } = useContext(TranslateContext);

  useEffect(() => {
    setLang(lang.replace(/[-]/, ""));
    setHanko((hanko) => {
      hanko.setLang(lang);
      return hanko;
    });
  }, [hanko, lang, setHanko, setLang]);

  return (
    <section part={"container"} className={styles.container} ref={ref}>
      {props.children}
    </section>
  );
});

export default Container;
