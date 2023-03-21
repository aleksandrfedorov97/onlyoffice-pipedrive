import AppExtensionsSDK, { Command } from "@pipedrive/app-extensions-sdk";
import axios from "axios";
import axiosRetry from "axios-retry";

import { ConfigResponse } from "src/types/config";

export const fetchConfig = async (
  sdk: AppExtensionsSDK,
  id: string,
  name: string,
  key: string,
  dealID: string,
  signal?: AbortSignal
) => {
  const pctx = await sdk.execute(Command.GET_SIGNED_TOKEN);
  const client = axios.create();
  axiosRetry(client, {
    retries: 2,
    retryCondition: (error) => error.status !== 200,
  });
  const res = await axios<ConfigResponse>({
    method: "GET",
    url: `${process.env.BACKEND_GATEWAY}/api/config`,
    params: {
      id,
      name,
      key,
      deal_id: dealID,
    },
    headers: {
      "Content-Type": "application/json",
      "X-Pipedrive-App-Context": pctx.token,
    },
    signal,
  });
  return res.data;
};
