import { useMutation, useQueryClient } from "react-query";

import { deleteFile } from "@services/file";

export const useDeleteFile = (url: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => deleteFile(url),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["filesData"] });
    },
  });
};
