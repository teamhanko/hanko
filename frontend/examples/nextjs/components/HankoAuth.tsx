import { register } from "@teamhanko/hanko-elements";
import { useCallback, useEffect } from "react";
import { useRouter } from "next/router";

const api = process.env.NEXT_PUBLIC_HANKO_API!;

interface Props {
  setError(error: Error): void;
}

function HankoAuth({ setError }: Props) {
  const router = useRouter();

  const redirectToTodos = useCallback(() => {
    router.replace("/todo").catch(setError);
  }, [router, setError]);

  useEffect(() => {
    register({ shadow: true }).catch(setError);
  }, [setError]);

  useEffect(() => {
    document.addEventListener("hankoAuthSuccess", redirectToTodos);
    return () =>
      document.removeEventListener("hankoAuthSuccess", redirectToTodos);
  }, [redirectToTodos]);

  return <hanko-auth api={api} />;
}

export default HankoAuth;
