import { h, InputHTMLAttributes } from "preact";
import {
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "preact/compat";

import styles from "./styles.sass";
import { AppContext } from "../../contexts/AppProvider";

// Inspired by https://github.com/devfolioco/react-otp-input

interface Props {
  passcodeDigits: string[];
  numberOfInputs?: number;
  onInput?: (passcodeDigits: string[]) => void;
  disabled?: boolean;
}

interface DigitProps extends InputHTMLAttributes {
  index: number;
  focus: boolean;
  digit: string;
}

const Digit = ({ index, focus, digit = "", ...props }: DigitProps) => {
  const ref = useRef(null);
  const { uiState } = useContext(AppContext);

  const focusInput = () => {
    const { current: element } = ref;
    if (element) {
      element.focus();
      element.select();
    }
  };

  const disabled = useMemo(
    () => uiState.isDisabled || props.disabled,
    [props, uiState],
  );

  // Autofocus if it's the first input element
  useEffect(() => {
    if (index === 0) {
      focusInput();
    }
  }, [index, props.disabled]);

  // Focus the current input element
  useMemo(() => {
    if (focus) {
      focusInput();
    }
  }, [focus]);

  return (
    <div className={styles.passcodeDigitWrapper}>
      <input
        {...props}
        part={"input passcode-input"}
        aria-label={`${props.name}-digit-${index + 1}`}
        name={props.name + index.toString(10)}
        type={"text"}
        inputMode={"numeric"}
        ref={ref}
        value={digit.charAt(0)}
        required={true}
        className={styles.input}
        disabled={disabled}
      />
    </div>
  );
};

const CodeInput = ({
  passcodeDigits = [],
  numberOfInputs = 6,
  onInput,
  disabled = false,
}: Props) => {
  const [activeInputIndex, setActiveInputIndex] = useState<number>(0);

  // returns a copy of the digit array
  const getPasscodeDigits = (): string[] => passcodeDigits.slice();

  const focusNextInput = () => {
    if (activeInputIndex < numberOfInputs - 1) {
      setActiveInputIndex(activeInputIndex + 1);
    }
  };

  const focusPrevInput = () => {
    if (activeInputIndex > 0) {
      setActiveInputIndex(activeInputIndex - 1);
    }
  };

  const changeCodeAtFocus = (digit: string) => {
    const digits = getPasscodeDigits();
    digits[activeInputIndex] = digit.charAt(0);
    onInput(digits);
  };

  // Writes the code to the inputs starting at the given position and focuses the input following the last written
  // digit, or the last input if the code reaches it
  const fillDigits = (code: string, startIndex: number) => {
    const digits = getPasscodeDigits();
    const chars = code.slice(0, numberOfInputs - startIndex).split("");

    chars.forEach((char, index) => {
      digits[startIndex + index] = char;
    });

    setActiveInputIndex(
      Math.min(startIndex + chars.length, numberOfInputs - 1),
    );
    onInput(digits);
  };

  const handleOnPaste = (event: ClipboardEvent) => {
    event.preventDefault();
    if (disabled) {
      return;
    }

    fillDigits(event.clipboardData.getData("text/plain"), activeInputIndex);
  };

  // Handle cases of backspace, delete, left arrow, right arrow, space
  const handleOnKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Backspace") {
      event.preventDefault();
      changeCodeAtFocus("");
      focusPrevInput();
    } else if (event.key === "Delete") {
      event.preventDefault();
      changeCodeAtFocus("");
    } else if (event.key === "ArrowLeft") {
      event.preventDefault();
      focusPrevInput();
    } else if (event.key === "ArrowRight") {
      event.preventDefault();
      focusNextInput();
    } else if (
      event.key === " " ||
      event.key === "Spacebar" ||
      event.key === "Space"
    ) {
      event.preventDefault();
    }
  };

  // The content may not have changed, but some input took place hence change the focus
  const handleOnInput = (event: Event) => {
    if (event.target instanceof HTMLInputElement) {
      // Only the first input accepts more than one character, which happens when the browser autofills the whole code
      if (event.target.value.length > 1) {
        fillDigits(event.target.value, 0);
        return;
      }
      changeCodeAtFocus(event.target.value);
    }
    focusNextInput();
  };

  const handleOnFocus = (index: number) => {
    setActiveInputIndex(index);
  };

  // Autofocus the first input when passcode has been reset
  useEffect(() => {
    if (passcodeDigits.length === 0) {
      setActiveInputIndex(0);
    }
  }, [passcodeDigits]);

  return (
    <div className={styles.passcodeInputWrapper}>
      {Array.from(Array(numberOfInputs)).map((_, index) => (
        <Digit
          name={"passcode"}
          key={index}
          index={index}
          focus={activeInputIndex === index}
          digit={passcodeDigits[index]}
          autocomplete={index === 0 ? "one-time-code" : "off"}
          maxLength={index === 0 ? numberOfInputs : 1}
          onKeyDown={handleOnKeyDown}
          onInput={handleOnInput}
          onPaste={handleOnPaste}
          onFocus={() => handleOnFocus(index)}
          disabled={disabled}
        />
      ))}
    </div>
  );
};

export default CodeInput;
