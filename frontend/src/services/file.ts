import { AuthToken } from "@context/TokenContext";
import axios from "axios";

import { FileResponse } from "src/types/file";

export const fetchFiles = async (
  url: string,
  start = 0,
  limit = 50,
  signal: AbortSignal | undefined = undefined,
  sort = "add_time ASC"
) => {
  const res = await axios<FileResponse>({
    method: "GET",
    url,
    params: {
      start,
      limit,
      sort,
    },
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${AuthToken.access_token}`,
    },
    signal,
  });

  return {
    response: res.data.data,
    pagination: res.data.additional_data.pagination,
  };
};

export const deleteFile = async (
  url: string,
  signal: AbortSignal | undefined = undefined
) => {
  const res = await axios({
    method: "DELETE",
    url,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${AuthToken.access_token}`,
    },
    signal,
    timeout: 4500,
  });

  return res.status === 200;
};

export const uploadFile = async (url: string, deal: string, file: File) => {
  const form = new FormData();
  form.append("file", file);
  form.append("deal_id", deal);

  const res = await axios({
    method: "POST",
    url,
    headers: {
      Accept: "application/json",
      "Content-Type": "multipart/form-data",
      Authorization: `Bearer ${AuthToken.access_token}`,
    },
    data: form,
  });

  return res.data;
};
