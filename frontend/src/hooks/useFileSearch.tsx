import { useInfiniteQuery } from "react-query";

import { fetchFiles } from "@services/file";

export function useFileSearch(url: string, limit: number) {
  const {
    data,
    isLoading,
    error,
    refetch,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: ["filesData", url],
    queryFn: ({ signal, pageParam }) =>
      fetchFiles(url, pageParam, limit, signal),
    getNextPageParam: (lastPage) =>
      lastPage.pagination.more_items_in_collection
        ? lastPage.pagination.next_start
        : undefined,
    staleTime: 2000,
    cacheTime: 2500,
    refetchInterval: 2000,
  });

  return {
    files: data?.pages
      .map((page) => page.response)
      .filter(Boolean)
      .flat(),
    isLoading,
    error,
    refetch,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  };
}
