import AppExtensionsSDK from "@pipedrive/app-extensions-sdk";
import React, { useEffect } from "react";
import { proxy } from "valtio";

import { getMe } from "@services/me";

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
        timerID = setTimeout(async function update() {
          if (
            !AuthToken.error &&
            (!AuthToken.access_token || AuthToken.expires_at <= Date.now() - 1)
          ) {
            try {
              const token = await getMe(sdk);
              AuthToken.access_token = token.response.access_token;
              AuthToken.expires_at = token.response.expires_at;
            } catch {
              AuthToken.error = true;
            }
          }
          timerID = setTimeout(update, 1000);
        }, 1000);
      })
      .catch(() => null);

    return () => clearTimeout(timerID);
  }, []);
  return <TokenContext.Provider value>{children}</TokenContext.Provider>;
};
