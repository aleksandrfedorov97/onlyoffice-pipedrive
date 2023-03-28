import { AuthToken } from "@context/TokenContext";
import axios from "axios";
import { PipedriveSearchUsersResponse } from "src/types/user";

export const fetchUsers = async (
  url: string,
  signal: AbortSignal | undefined = undefined
) => {
  const res = await axios<PipedriveSearchUsersResponse>({
    method: "GET",
    url,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${AuthToken.access_token}`,
    },
    signal,
    timeout: 4000,
  });

  return {
    response: res.data,
  };
};
