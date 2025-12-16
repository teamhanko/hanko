import { ComponentChildren, Fragment, h } from "preact";
import { Dispatch, SetStateAction } from "preact/compat";

import Accordion from "./Accordion";

interface Props {
  name: string;
  title: string | h.JSX.Element;
  children: ComponentChildren;
  checkedItemID?: string;
  setCheckedItemID: Dispatch<SetStateAction<string>>;
}

const Dropdown = ({
  name,
  title,
  children,
  checkedItemID,
  setCheckedItemID,
}: Props) => {
  return (
    <Accordion
      dropdown
      name={name}
      columnSelector={() => title}
      contentSelector={() => <>{children}</>}
      setCheckedItemID={setCheckedItemID}
      checkedItemID={checkedItemID}
      data={[{}]}
    />
  );
};

export default Dropdown;
