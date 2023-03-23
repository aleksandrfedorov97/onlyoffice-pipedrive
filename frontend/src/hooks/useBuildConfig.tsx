import { useQuery } from "react-query";

import { fetchConfig } from "@services/config";

export function useBuildConfig(
  token: string,
  id: string,
  name: string,
  key: string,
  dealID: string
) {
  const { isLoading, error, data } = useQuery({
    queryKey: ["config", id],
    queryFn: ({ signal }) => fetchConfig(token, id, name, key, dealID, signal),
    staleTime: 0,
    cacheTime: 0,
    refetchOnWindowFocus: false,
  });

  return { isLoading, error, data };
}
