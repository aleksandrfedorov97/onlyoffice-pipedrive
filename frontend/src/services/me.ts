import axios from "axios";
import axiosRetry from "axios-retry";
import AppExtensionsSDK, { Command } from "@pipedrive/app-extensions-sdk";
import { PipedriveUserResponse, UserResponse } from "src/types/user";
import { AuthToken } from "@context/TokenContext";

export const getMe = async (sdk: AppExtensionsSDK) => {
  const pctx = await sdk.execute(Command.GET_SIGNED_TOKEN);
  const client = axios.create({ baseURL: process.env.BACKEND_GATEWAY });
  axiosRetry(client, {
    retries: 2,
    retryCondition: (error) => error.status !== 200,
    retryDelay: (count) => count * 50,
  });
  const res = await client<UserResponse>({
    method: "GET",
    url: `/api/me`,
    headers: {
      "Content-Type": "application/json",
      "X-Pipedrive-App-Context": pctx.token,
    },
  });

  return { response: res.data };
};

export const getPipedriveMe = async (url: string) => {
  const res = await axios<PipedriveUserResponse>({
    method: "GET",
    url,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${AuthToken.access_token}`,
    },
  });
  return res.data;
};
