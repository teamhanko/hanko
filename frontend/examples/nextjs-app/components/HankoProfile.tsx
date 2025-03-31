"use client";

import { register } from "@teamhanko/hanko-elements";
import { useEffect } from "react";

interface Props {
  setError(error: Error): void;
}

function HankoProfile({ setError }: Props) {
  const api = process.env.NEXT_PUBLIC_HANKO_API!;

  useEffect(() => {
    register(api).catch(setError);
  }, [setError, api]);

  return <hanko-profile />;
}

export default HankoProfile;
