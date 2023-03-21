import { useQuery } from "react-query";
import { useSnapshot } from "valtio";

import { fetchConfig } from "@services/config";

import { PipedriveSDK } from "@context/PipedriveContext";

export function useBuildConfig(
  id: string,
  name: string,
  key: string,
  dealID: string
) {
  const { sdk } = useSnapshot(PipedriveSDK);
  const { isLoading, error, data } = useQuery({
    queryKey: ["config", id],
    queryFn: ({ signal }) => fetchConfig(sdk, id, name, key, dealID, signal),
    staleTime: 0,
    cacheTime: 0,
    refetchOnWindowFocus: false,
  });

  return { isLoading, error, data };
}
