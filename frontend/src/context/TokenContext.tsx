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

import AppExtensionsSDK from "@pipedrive/app-extensions-sdk";
import i18next from "i18next";
import axios from "axios";
import React, { useEffect } from "react";
import { proxy } from "valtio";

import { getMe, getPipedriveMe } from "@services/me";

import { getCurrentURL } from "@utils/url";
import { getWithExpiry, setWithExpiry } from "@utils/storage";
import { UserResponse } from "src/types/user";

export const AuthToken = proxy({
  access_token: "",
  expires_at: Date.now(),
  error: false,
});

type ProviderProps = {
  children?: JSX.Element | JSX.Element[];
};

const TokenContext = React.createContext<boolean>(true);

export const TokenProvider: React.FC<ProviderProps> = ({ children }) => {
  useEffect(() => {
    let timerID: NodeJS.Timeout;
    new AppExtensionsSDK()
      .initialize()
      .then((sdk) => {
        const { url } = getCurrentURL();
        timerID = setTimeout(async function update() {
          if (
            !AuthToken.error &&
            (!AuthToken.access_token ||
              AuthToken.expires_at <= Date.now() - 1000 * 19)
          ) {
            try {
              const val = getWithExpiry("authorization") as UserResponse;
              if (val) {
                AuthToken.access_token = val.access_token;
                AuthToken.expires_at = val.expires_at;
              } else {
                const token = await getMe(sdk);
                AuthToken.access_token = token.response.access_token;
                AuthToken.expires_at = token.response.expires_at;
                setWithExpiry(
                  "authorization",
                  token.response,
                  token.response.expires_at
                );
              }
              const resp = await getPipedriveMe(`${url}api/v1/users/me`);
              await i18next.changeLanguage(resp.data.language.language_code);
            } catch (err) {
              if (!axios.isCancel(err)) {
                AuthToken.error = true;
                AuthToken.access_token = "";
              }
            }
          }
          timerID = setTimeout(
            update,
            AuthToken.expires_at - Date.now() - 1000 * 19
          );
        }, 0);
      })
      .catch(() => null);

    return () => clearTimeout(timerID);
  }, []);
  return <TokenContext.Provider value>{children}</TokenContext.Provider>;
};
