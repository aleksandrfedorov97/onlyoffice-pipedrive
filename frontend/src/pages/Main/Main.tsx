import React, { useCallback, useEffect, useRef, useState } from "react";
import AppExtensionsSDK, {
  Command,
  Modal,
} from "@pipedrive/app-extensions-sdk";

import { OnlyofficeButton } from "@components/button";
import { OnlyofficeFile } from "@components/file";
import { OnlyofficeFileInfo } from "@components/info";
import { OnlyofficeNoFile } from "@components/nofile";
import { OnlyofficeSpinner } from "@components/spinner";

import { useFileSearch } from "@hooks/useFileSearch";

import { formatBytes, getFileIcon, isFileSupported } from "@utils/file";
import { getCurrentURL } from "@utils/url";

import { OnlyofficeFileActions } from "./Actions";

export const Main: React.FC = () => {
  const { url, parameters } = getCurrentURL();
  const [sdk, setSDK] = useState<AppExtensionsSDK | null>();
  const { isLoading, fetchNextPage, isFetchingNextPage, files, hasNextPage } =
    useFileSearch(
      `${url}api/v1/deals/${parameters.get("selectedIds")}/files`,
      20
    );

  const observer = useRef<IntersectionObserver>();
  const lastItem = useCallback(
    (node: Element | null) => {
      if (isLoading) return;
      if (observer.current) observer.current.disconnect();
      observer.current = new IntersectionObserver(async (entries) => {
        if (entries[0].isIntersecting && hasNextPage) {
          fetchNextPage();
        }
      });
      if (node) observer.current.observe(node);
    },
    [isLoading, fetchNextPage, hasNextPage]
  );

  useEffect(() => {
    new AppExtensionsSDK()
      .initialize()
      .then((s) => setSDK(s))
      .catch(() => setSDK(null));
  }, []);

  return (
    <div className="table-shadow h-full">
      <div className="px-5 overflow-scroll h-[85%] md:justify-between no-scrollbar">
        {isLoading && (
          <div className="h-[85%] w-full flex justify-center items-center">
            <OnlyofficeSpinner />
          </div>
        )}
        {!isLoading && (!files || files.length === 0) && (
          <OnlyofficeNoFile title="Could not find pipedrive files" />
        )}
        {!isLoading &&
          files &&
          files.length > 0 &&
          files.map((file, index) => {
            if (files.length === index + 1) {
              return (
                <div key={file.id + file.add_time} ref={lastItem}>
                  <OnlyofficeFile
                    Icon={getFileIcon(file.name)}
                    name={file.name}
                    supported={isFileSupported(file.name)}
                    actions={<OnlyofficeFileActions file={file} />}
                  >
                    <OnlyofficeFileInfo
                      info={{
                        // "Created by": file.person_name,
                        Workspace: file.remote_location,
                        Type: file.file_type,
                        "Date modified": file.update_time,
                        "Creation date": file.add_time,
                        Size: formatBytes(file.file_size),
                      }}
                    />
                  </OnlyofficeFile>
                </div>
              );
            }
            return (
              <div key={file.id + file.add_time}>
                <OnlyofficeFile
                  Icon={getFileIcon(file.name)}
                  name={file.name}
                  supported={isFileSupported(file.name)}
                  actions={<OnlyofficeFileActions file={file} />}
                >
                  <OnlyofficeFileInfo
                    info={{
                      // "Created by": file.person_name,
                      Workspace: file.remote_location,
                      Type: file.file_type,
                      "Date modified": file.update_time,
                      "Creation date": file.add_time,
                      Size: formatBytes(file.file_size),
                    }}
                  />
                </OnlyofficeFile>
              </div>
            );
          })}
        {isFetchingNextPage && (
          <div
            className={`relative w-full ${
              isLoading ? "h-full" : "h-fit"
            } my-5 flex justify-center items-center`}
          >
            <OnlyofficeSpinner />
          </div>
        )}
      </div>
      <div className="h-[15%] w-3/4 text-ellipsis flex justify-center items-center px-5">
        <OnlyofficeButton
          text="Create or upload document"
          fullWidth
          primary
          onClick={async () => {
            await sdk?.execute(Command.OPEN_MODAL, {
              type: Modal.CUSTOM_MODAL,
              action_id: process.env.PIPEDRIVE_CREATE_MODAL_ID || "",
            });
          }}
        />
      </div>
    </div>
  );
};
