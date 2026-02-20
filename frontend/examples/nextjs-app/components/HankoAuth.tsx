"use client";

import { Hanko, register } from "@teamhanko/hanko-elements";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";

interface Props {
  setError(error: Error): void;
}

function HankoAuth({ setError }: Props) {
  const router = useRouter();
  const [hankoClient, setHankoClient] = useState<Hanko>();
  const api = process.env.NEXT_PUBLIC_HANKO_API!;

  const redirectToTodos = useCallback(() => {
    router.push("/todo");
  }, [router]);

  useEffect(() => {
    setHankoClient(new Hanko(api));
  }, [api]);

  useEffect(() => {
    register(api).catch(setError);
  }, [setError, api]);

  useEffect(() => {
    if (hankoClient) {
      hankoClient.onSessionCreated(() => {
        redirectToTodos();
      });
    }
  }, [hankoClient, redirectToTodos]);

  return <hanko-auth />;
}

export default HankoAuth;
