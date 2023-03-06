import { h } from "preact";
import { StateUpdater } from "preact/compat";

import cx from "classnames";

import styles from "./styles.sass";

type Selector<T> = (item: T, itemIndex?: number) => string | h.JSX.Element;

interface Props<T> {
  name: string;
  columnSelector: Selector<T>;
  contentSelector: Selector<T>;
  checkedItemIndex?: number;
  setCheckedItemIndex: StateUpdater<number>;
  data: Array<T>;
  dropdown?: boolean;
}

const Accordion = function <T>({
  name,
  columnSelector,
  contentSelector,
  data,
  checkedItemIndex,
  setCheckedItemIndex,
  dropdown = false,
}: Props<T>) {
  const clickHandler = (event: Event) => {
    if (!(event.target instanceof HTMLInputElement)) return;
    const itemIndex = parseInt(event.target.value, 10);
    setCheckedItemIndex(itemIndex === checkedItemIndex ? null : itemIndex);
  };

  return (
    <div className={styles.accordion}>
      {data.map((item, itemIndex) => (
        <div className={styles.accordionItem} key={itemIndex}>
          <input
            type={"radio"}
            className={styles.accordionInput}
            id={`${name}-${itemIndex}`}
            name={name}
            onClick={clickHandler}
            value={itemIndex}
            checked={checkedItemIndex === itemIndex}
          />
          <label
            className={cx(styles.label, dropdown && styles.dropdown)}
            for={`${name}-${itemIndex}`}
          >
            <span className={styles.labelText}>
              {columnSelector(item, itemIndex)}
            </span>
          </label>
          <div
            className={cx(
              styles.accordionContent,
              dropdown && styles.dropdownContent
            )}
          >
            {contentSelector(item, itemIndex)}
          </div>
        </div>
      ))}
    </div>
  );
};

export default Accordion;
