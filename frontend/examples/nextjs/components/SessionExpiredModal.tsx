import { forwardRef, useCallback } from "react";
import { useRouter } from "next/router";

export const SessionExpiredModal = forwardRef<HTMLDialogElement>(
  (props, ref) => {
    const router = useRouter();

    const redirectToLogin = useCallback(() => {
      router.push("/");
    }, [router]);

    return (
      <dialog ref={ref}>
        Please login again.
        <br />
        <br />
        <button onClick={redirectToLogin}>Login</button>
      </dialog>
    );
  }
);

SessionExpiredModal.displayName = "SessionExpiredModal"
