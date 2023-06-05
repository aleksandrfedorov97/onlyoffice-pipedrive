/**
 *
 * (c) Copyright Ascensio System SIA 2023
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import React from "react";
import { useSnapshot } from "valtio";
import { useTranslation } from "react-i18next";

import { OnlyofficeSpinner } from "@components/spinner";

import { OnlyofficeBackgroundError } from "@layouts/ErrorBackground";

import { AuthToken } from "@context/TokenContext";

import TokenError from "@assets/token-error.svg";

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
          Icon={<TokenError className="mb-5" />}
          title={t(
            "background.reinstall.title",
            "The document security token has expired"
          )}
          subtitle={t(
            "background.reinstall.subtitle",
            "Something went wrong. Please reload or reinstall the app."
          )}
          button={t(
            "background.reinstall.button",
            "Reinstall"
          ) || "Reinstall"}
          onClick={() => window.open(`${process.env.BACKEND_GATEWAY}/oauth/install`, "_blank")}
        />
      )}
      {loaded && <Main />}
    </div>
  );
};
