import React from "react";
import { useSnapshot } from "valtio";
import { useTranslation } from "react-i18next";

import { OnlyofficeSpinner } from "@components/spinner";

import { OnlyofficeBackgroundError } from "@layouts/ErrorBackground";

import { AuthToken } from "@context/TokenContext";

import BackgroundError from "@assets/background-error.svg";

import { Main } from "./Main";

export const MainPage: React.FC = () => {
  const { t } = useTranslation();
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
          Icon={<BackgroundError />}
          title={t("background.error.title", "Error")}
          subtitle={t(
            "background.reinstall.subtitle",
            "Something went wrong. Please reload or reinstall the app."
          )}
        />
      )}
      {loaded && <Main />}
    </div>
  );
};
