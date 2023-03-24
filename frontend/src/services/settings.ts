import axios from "axios";
import axiosRetry from "axios-retry";
import AppExtensionsSDK, { Command } from "@pipedrive/app-extensions-sdk";
import { SettingsResponse } from "src/types/settings";

export const postSettings = async (
  sdk: AppExtensionsSDK,
  address: string,
  secret: string,
  header: string
) => {
  const pctx = await sdk.execute(Command.GET_SIGNED_TOKEN);
  const client = axios.create({ baseURL: process.env.BACKEND_GATEWAY });
  axiosRetry(client, {
    retries: 2,
    retryCondition: (error) => error.status === 429,
  });

  await client({
    method: "POST",
    url: `/api/settings`,
    headers: {
      "Content-Type": "application/json",
      "X-Pipedrive-App-Context": pctx.token,
    },
    data: {
      doc_address: address,
      doc_secret: secret,
      doc_header: header,
    },
    timeout: 4000,
  });
};

export const getSettings = async (sdk: AppExtensionsSDK) => {
  const pctx = await sdk.execute(Command.GET_SIGNED_TOKEN);
  const client = axios.create({ baseURL: process.env.BACKEND_GATEWAY });
  axiosRetry(client, {
    retries: 2,
    retryCondition: (error) => error.status !== 200,
  });

  const settings = await client<SettingsResponse>({
    method: "GET",
    url: `/api/settings`,
    headers: {
      "Content-Type": "application/json",
      "X-Pipedrive-App-Context": pctx.token,
    },
  });

  return settings.data;
};
