import { forwardRef, useCallback } from "react";
import { useNavigate } from "react-router-dom";

export const SessionExpiredModal = forwardRef<HTMLDialogElement>(
  (props, ref) => {
    const navigate = useNavigate();

    const redirectToLogin = useCallback(() => {
      navigate("/");
    }, [navigate]);

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
