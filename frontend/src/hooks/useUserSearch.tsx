import { useQuery } from "react-query";
import { fetchUsers } from "@services/user";

export function useUserSearch(url: string) {
  const { isLoading, error, data } = useQuery({
    queryKey: ["users"],
    queryFn: ({ signal }) => fetchUsers(url, signal),
    staleTime: 30000,
    cacheTime: 30000,
    refetchOnWindowFocus: false,
  });

  return { isLoading, error, data };
}
