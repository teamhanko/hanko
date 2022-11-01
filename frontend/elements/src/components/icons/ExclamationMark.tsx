import * as preact from "preact";

import styles from "./styles.sass";

const ExclamationMark = () => {
  return (
    <div className={styles.exclamationMark}>
      <div className={styles.circle} />
      <div className={styles.stem} />
      <div className={styles.dot} />
    </div>
  );
};

export default ExclamationMark;
