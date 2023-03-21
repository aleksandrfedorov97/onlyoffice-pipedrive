import React from "react";
import { useSnapshot } from "valtio";

import { OnlyofficeSpinner } from "@components/spinner";

import { AuthToken } from "@context/TokenContext";

import { OnlyofficeBackgroundError } from "@layouts/ErrorBackground";
import { Main } from "./Main";

export const MainPage: React.FC = () => {
  const { access_token: accessToken, error } = useSnapshot(AuthToken);
  const loading = !accessToken && !error;
  const loadingError = !accessToken && error;
  const loaded = accessToken && !error;
  return (
    <div className="relative w-full h-full flex flex-col my-0 mx-auto">
      {loading && (
        <div className="w-full h-full flex justify-center items-center">
          <OnlyofficeSpinner />
        </div>
      )}
      {loadingError && (
        <OnlyofficeBackgroundError
          title="Error"
          subtitle="Something went wrong. Please reload or reinstall the app."
        />
      )}
      {loaded && <Main />}
    </div>
  );
};
