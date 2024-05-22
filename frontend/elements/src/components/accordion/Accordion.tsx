import { h } from "preact";
import { StateUpdater, useCallback } from "preact/compat";
type Selector<T> = (item: T, itemIndex?: number) => string | h.JSX.Element;

import cx from "classnames";
import styles from "./styles.sass";

interface Props<T> {
  name: string;
  columnSelector: Selector<T>;
  contentSelector: Selector<T>;
  checkedItemID?: string;
  setCheckedItemID: StateUpdater<string>;
  data: Array<T>;
  dropdown?: boolean;
}

const Accordion = function <T>({
  name,
  columnSelector,
  contentSelector,
  data,
  checkedItemID,
  setCheckedItemID,
  dropdown = false,
}: Props<T>) {
  const toID = useCallback(
    (itemIndex: number) => `${name}-${itemIndex}`,
    [name],
  );

  const checked = useCallback(
    (itemIndex: number) => toID(itemIndex) === checkedItemID,
    [checkedItemID, toID],
  );

  const clickHandler = (event: Event) => {
    if (!(event.target instanceof HTMLInputElement)) return;
    const itemIndex = parseInt(event.target.value, 10);
    const id = toID(itemIndex);
    setCheckedItemID(id === checkedItemID ? null : id);
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
            checked={checked(itemIndex)}
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
              dropdown && styles.dropdownContent,
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
