import { useContext, useEffect } from "preact/compat";

import { AppContext } from "../contexts/AppProvider";

import LoadingSpinner from "../components/icons/LoadingSpinner";

const InitPage = () => {
  const { setLoadingAction } = useContext(AppContext);

  useEffect(() => {
    setLoadingAction(null);
  }, []);

  return <LoadingSpinner isLoading />;
};

export default InitPage;
