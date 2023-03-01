import * as preact from "preact";
import * as icons from "./icons";

export type IconName = keyof typeof icons;

export type IconProps = {
  secondary?: boolean;
  fadeOut?: boolean;
  disabled?: boolean;
  size?: number;
  style?: string;
};

type Props = IconProps & {
  name: IconName;
};

const Icon = ({
  name,
  secondary,
  size = 18,
  fadeOut,
  style,
  disabled,
}: Props) => {
  const Ico = icons[name];

  return (
    <Ico
      size={size}
      secondary={secondary}
      fadeOut={fadeOut}
      style={style}
      disabled={disabled}
    />
  );
};

export default Icon;
