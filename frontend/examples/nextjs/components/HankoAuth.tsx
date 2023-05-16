import { Hanko, register } from "@teamhanko/hanko-elements";
import { useCallback, useEffect, useMemo } from "react";
import { useRouter } from "next/router";

const api = process.env.NEXT_PUBLIC_HANKO_API!;

interface Props {
  setError(error: Error): void;
}

function HankoAuth({ setError }: Props) {
  const router = useRouter();
  const hankoClient = useMemo(() => new Hanko(api), []);

  const redirectToTodos = useCallback(() => {
    router.replace("/todo").catch(setError);
  }, [router, setError]);

  useEffect(() => {
    register(api).catch(setError);
  }, [setError]);

  useEffect(() => hankoClient.onAuthFlowCompleted(() => {
    redirectToTodos()
  }), [hankoClient, redirectToTodos]);

  return <hanko-auth />;
}

export default HankoAuth;
